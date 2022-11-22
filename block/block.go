package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"time"
)

//将int型数据转行为16进制
func IntToHex(data int64) []byte {
	buffer := new(bytes.Buffer) // 新建一个buffer
	err := binary.Write(buffer, binary.BigEndian, data)
	if nil != err {
		log.Panicf("int to []byte failed! %v\n", err)
	}
	return buffer.Bytes()
}

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBit))
	pow := &ProofOfWork{b, target}
	return pow
}

//Nonce和区块其他信息一起进行Hash计算
func (pow *ProofOfWork) prepareData(nonce int64) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.Index),
			IntToHex(pow.block.TimeStamp),
			IntToHex(int64(targetBit)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}
func (pow *ProofOfWork) Run() (int64, []byte) {
	var hashInt big.Int
	var hash [32]byte
	var nonce int64 = 0
	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	for {
		dataBytes := pow.prepareData(nonce) //获取准备的数据
		hash = sha256.Sum256(dataBytes)     //对数据进行Hash
		hashInt.SetBytes(hash[:])
		fmt.Printf("hash: \r%x", hash)
		if pow.target.Cmp(&hashInt) == 1 { //对比hash值
			break
		}
		nonce++ //充当计数器，同时在循环结束后也是符合要求的值

	}
	fmt.Printf("\n碰撞次数: %d\n", nonce)
	return int64(nonce), hash[:]
}

const targetBit = 20

type Block struct {
	Index         int64
	TimeStamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int64 //This line is new
}
type Blockchain struct {
	blocks []*Block
}

//保存Hash结果在当前块的Hash中
func (b *Block) setHash() {
	timestamp := []byte(strconv.FormatInt(b.TimeStamp, 10))
	Nonce := []byte(strconv.FormatInt(b.Nonce, 10))
	headers := bytes.Join([][]byte{timestamp, Nonce, b.Data, b.PrevBlockHash}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:] //保存Hash结果在当前块的Hash中
}

//构建Block新对象
func NewBlock(index int64, data []byte, prevBlockHash []byte) *Block {
	block := &Block{index, time.Now().Unix(), data, prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

//创建创世区块
func NewGenesisBlock() *Block {
	return NewBlock(0, []byte("first block"), []byte{})
}

//将创世区块入链
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock()}}
}

//添加新区块到链
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(prevBlock.Index+1, []byte(data), prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}
func main() {
	bc := NewBlockchain()
	fmt.Printf("blockChain : %v\n", bc)
	bc.AddBlock("Aimi send 100 BTC	to Bob")
	bc.AddBlock("Aimi send 100 BTC	to Jay")
	bc.AddBlock("Aimi send 100 BTC	to Clown")
	length := len(bc.blocks)
	fmt.Printf("length of blocks : %d\n", length)
	for i := 0; i < length; i++ {
		pow := NewProofOfWork(bc.blocks[i])
		if pow.Validate() {
			fmt.Println("—————————————————————————————————————————————————————")
			fmt.Printf(" Block: %d\n", bc.blocks[i].Index)
			fmt.Printf("Data: %s\n", bc.blocks[i].Data)
			fmt.Printf("TimeStamp: %d\n", bc.blocks[i].TimeStamp)
			fmt.Printf("Hash: %x\n", bc.blocks[i].Hash)
			fmt.Printf("PrevHash: %x\n", bc.blocks[i].PrevBlockHash)
			fmt.Printf("Nonce: %d\n", bc.blocks[i].Nonce)

		} else {
			fmt.Println("illegal block")
		}
	}
}
