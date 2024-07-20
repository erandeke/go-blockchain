package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Book struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	PublishDate string `json:"publishedDate"`
	ISBN        string `json:"isbn"`
}

type Blocks struct {
	Pos       int
	Data      BookCheckOut
	Timestamp string
	Prevhash  string
	Hash      string
}

type BlockChain struct {
	blocks []*Blocks
}

type BookCheckOut struct {
	BookId       string `json:"bookId"`
	User         string `json:"user"`
	checkoutDate string `json:"checkoutDate"`
	IsGenesis    bool   `json:"is_genesis"`
}

var blockchain *BlockChain

func NewBlockchain() *BlockChain {
	return &BlockChain{[]*Blocks{GenesisBlock()}}
}

func GenesisBlock() *Blocks {
	return CreateNewBlock(&Blocks{}, BookCheckOut{IsGenesis: true})

}

func main() {
	blockchain := NewBlockchain()

	log.Print("Starting the blockchain instance")
	r := mux.NewRouter()
	r.HandleFunc("/", getTheWholeBlockChain).Methods("GET")
	r.HandleFunc("/new", createNewBook).Methods("POST")
	r.HandleFunc("/writeBlock", writeBlock).Methods("POST")

	go func() {

		for _, block := range blockchain.blocks {

			fmt.Printf("The prev hash is %v\n", block.Prevhash)
			fmt.Printf("The pos  is %v\n", block.Pos)
			fmt.Printf("The  hash is %v\n", block.Hash)
			bytes, _ := json.MarshalIndent(block.Data, "", "")
			fmt.Printf("Data is %v\n", string(bytes))

		}

	}()

	log.Fatal(http.ListenAndServe(":8080", r))

}

func createNewBook(w http.ResponseWriter, r *http.Request) {

	var book Book

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not decode the incoming the book data"))
		return
	}

	id := md5.New()
	io.WriteString(id, book.ISBN+book.PublishDate)
	book.ID = fmt.Sprintf("%x", id.Sum(nil))

	resp, err := json.MarshalIndent(book, "", "")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not save the book"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp) // Write the data to Http as the part of reply

}

func writeBlock(w http.ResponseWriter, r *http.Request) {
	var bookCheckOut BookCheckOut
	if err := json.NewDecoder(r.Body).Decode(&bookCheckOut); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not decode the incoming the book data"))
		return

	}

	blockchain.AddBlocks(bookCheckOut)
}

func (b *BlockChain) AddBlocks(data BookCheckOut) {

	// assign prev block
	prevBlock := b.blocks[len(b.blocks)-1]

	//createv a new block
	block := CreateNewBlock(prevBlock, data)

	//validate the createdBlock
	if validatBlock(block, prevBlock) {
		b.blocks = append(b.blocks, block)

	}

}

func CreateNewBlock(prevBlock *Blocks, data BookCheckOut) *Blocks {
	var block Blocks
	block.Prevhash = prevBlock.Prevhash
	block.Pos = prevBlock.Pos + 1
	block.Data = data
	block.Timestamp = time.Now().String()
	block.generateHash()

	return &block

}

func validatBlock(createdBlock, prevBlock *Blocks) bool {

	if prevBlock.Hash != createdBlock.Hash {
		return false
	}

	if prevBlock.Pos+1 != createdBlock.Pos {
		return false
	}

	if !createdBlock.validateHash(createdBlock.Hash) {
		return false

	}

	return true

}

func (b *Blocks) validateHash(createdBlocksHash string) bool {
	b.generateHash()
	return b.Hash == createdBlocksHash
}

func (b *Blocks) generateHash() {

	//get the data and generate the hash using sha256

	bytes, _ := json.Marshal(b.Data)

	//prepare the data
	data := string(b.Pos) + b.Timestamp + b.Prevhash + string(bytes)

	hash := sha256.Sum256([]byte(data))

	b.Hash = hex.EncodeToString(hash[:])

}

func getTheWholeBlockChain(w http.ResponseWriter, r *http.Request) {

	bytes, err := json.MarshalIndent(blockchain.blocks, "", "")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not fetch the blockchains"))
	}

	w.Write(bytes)

}
