package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
    "fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)
type Block struct {
	Index int
	timestamp string
	BPM int
	hash string
	prevHash string
}
var Blockchain []Block

type Message struct {
	BPM int
}
//to calculate hash of new block
func calcHash(block Block) string {
	record:=string(block.Index)+block.timestamp + string(block.BPM) + block.prevHash
	h:=sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

//to generate new Block
func generateBlock(oldBlock Block, BPM int) (Block, error) {
	
		var newBlock Block
	
		t := time.Now()
	
		newBlock.Index = oldBlock.Index + 1
		newBlock.timestamp = t.String()
		newBlock.BPM = BPM
		newBlock.prevHash = oldBlock.hash
		newBlock.hash = calcHash(newBlock)
	
		return newBlock, nil
}
//to check that new block added  is valid
func validate(old,current Block) bool {

	if old.Index+1!=current.Index {
		return false
	} else if old.hash!=current.prevHash  {
		return false
	} else if calcHash(current)!=current.hash {
		return false
	} else {
		return true
	}
}

//PoL proof of length .Miner who is succesful in generating the maximum chain length gets to add the block
func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}

func runServer() error {
	mux := makeMuxRouter()
	httpAddr := os.Getenv("ADDR")
	log.Println("Listening on ", os.Getenv("ADDR"))
	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil

}
//routeHandlers
func getChain(w http.ResponseWriter,r* http.Request) {
	fmt.Println(Blockchain);
	bytes, err := json.MarshalIndent(Blockchain,"", "    ")
	if err!=nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w,string(bytes))
}
func writeBlock(w http.ResponseWriter,r* http.Request) {
  
	newBlock,err:=generateBlock(Blockchain[len(Blockchain)-1],50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if validate(Blockchain[len(Blockchain)-1],newBlock) {
		newBlockchain:=append(Blockchain,newBlock)
		replaceChain(newBlockchain)
		spew.Dump(Blockchain)
	}
}
/////////
//route handler
func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", getChain).Methods("GET")
	muxRouter.HandleFunc("/write", writeBlock).Methods("POST")
	return muxRouter
}
func main() {
	err := godotenv.Load()
    if err!=nil {
		log.Fatal(err)
	}	
	go func() {
		t := time.Now()
		genesisBlock := Block{0, t.String(), 0, "", ""}
		spew.Dump(genesisBlock)
		Blockchain = append(Blockchain, genesisBlock)
	}()
	log.Fatal(runServer())
}