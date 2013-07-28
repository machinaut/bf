package bf

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
)

const (
	NEXT = byte('>')
	PREV = byte('<')
	INC  = byte('+')
	DEC  = byte('-')
	PUT  = byte('.')
	GET  = byte(',')
	JUMP = byte('[')
	LOOP = byte(']')
	MOVE = byte('{') // compiled NEXT
	NOVE = byte('}') // compiled PREV (a "negative" move)
	MATH = byte('&') // compile INC
	NATH = byte('|') // compile DEC (a "negative" math)
)

// brainfuck interpreter
type BFInterp struct {
	mem   [30000]byte   // memory
	ptr   int           // pointer into memory
	pc    int           // index in program string (processor counter)
	prog  []byte        // program
	in    *bufio.Reader // input source
	out   *bufio.Writer // output sink
	jumps map[int]int   // jump map (left braces '[')
	loops map[int]int   // loop map (inverse of jump map; right braces ']')
	moves map[int]int   // map compressed moves
	noves map[int]int   // map compressed noves
	maths map[int]int   // map compressed maths
	naths map[int]int   // map compressed naths
}

func NewBFInterp(prog io.Reader, in io.Reader, out io.Writer) *BFInterp {
	bfi := new(BFInterp)
	bfi.prog, _ = ioutil.ReadAll(prog)
	bfi.in = bufio.NewReader(in)
	bfi.out = bufio.NewWriter(out)
	bfi.jumps = make(map[int]int)
	bfi.loops = make(map[int]int)
	bfi.moves = make(map[int]int)
	bfi.noves = make(map[int]int)
	bfi.maths = make(map[int]int)
	bfi.naths = make(map[int]int)

	return bfi
}

// Run the program
func (bfi *BFInterp) Run() {
	// preprocess the program
	bfi.scan()

	//bfi.Dump()

	fmt.Println("Program length", len(bfi.prog))

	for bfi.pc < len(bfi.prog) {
		bfi.Step()
	}

	bfi.out.Flush()
}

// Run a single instruction
func (bfi *BFInterp) Step() {
	cmd := bfi.prog[bfi.pc]
	//fmt.Printf("pc:%d %c ", bfi.pc, cmd)
	switch cmd {
	case NEXT:
		bfi.Next()
	case PREV:
		bfi.Prev()
	case INC:
		bfi.Inc()
	case DEC:
		bfi.Dec()
	case PUT:
		bfi.Put()
	case GET:
		bfi.Get()
	case JUMP:
		bfi.Jump()
	case LOOP:
		bfi.Loop()
	case MOVE:
		bfi.Move()
	case NOVE:
		bfi.Nove()
	case MATH:
		bfi.Math()
	case NATH:
		bfi.Nath()
	default:
		bfi.pc += 1
	}
}

// Scan and find matching brace sets
func (bfi *BFInterp) scan() {
	// quick-n-dirty stack made out of a slice
	var stack []int

	for idx, cmd := range bfi.prog {
		// TODO: clean this up.  This shit is nasty.
		if cmd == INC || cmd == DEC || cmd == NEXT || cmd == NEXT {
			// scan forward to find length of the 'run'
			var i int
			for i = 0; i+idx < len(bfi.prog) && bfi.prog[i+idx] == cmd; i++ {
				// TODO: theres probably a better way to do this :P
			}
			if i > 3 { // MAGIC NUMBER for optimizing
				// Overwrite command
				switch cmd {
				case NEXT:
					bfi.prog[idx] = MOVE
					bfi.moves[idx] = i
				case PREV:
					bfi.prog[idx] = NOVE
					bfi.noves[idx] = i
				case INC:
					bfi.prog[idx] = MATH
					bfi.maths[idx] = i
				case DEC:
					bfi.prog[idx] = NATH
					bfi.naths[idx] = i
				}
			}
		}
		if cmd == JUMP {
			stack = append(stack, idx) // push
		} else if cmd == LOOP {
			x := stack[len(stack)-1] // read last item
			bfi.jumps[x] = idx
			bfi.loops[idx] = x
			stack = stack[:len(stack)-1] // pop
		}
	}
}

// Dump table info to stdout for debugging
func (bfi *BFInterp) Dump() {
	fmt.Println("Jumps:")
	fmt.Println(bfi.jumps)
	fmt.Println("Loops:")
	fmt.Println(bfi.loops)
	fmt.Println("Moves:")
	fmt.Println(bfi.moves)
	fmt.Println("Noves:")
	fmt.Println(bfi.noves)
	fmt.Println("Maths:")
	fmt.Println(bfi.maths)
	fmt.Println("Naths:")
	fmt.Println(bfi.naths)
}

// >
func (bfi *BFInterp) Next() {
	bfi.ptr += 1
	bfi.pc += 1
}

// <
func (bfi *BFInterp) Prev() {
	bfi.ptr -= 1
	bfi.pc += 1
}

// +
func (bfi *BFInterp) Inc() {
	bfi.mem[bfi.ptr] += 1
	bfi.pc += 1
}

// -
func (bfi *BFInterp) Dec() {
	bfi.mem[bfi.ptr] -= 1
	bfi.pc += 1
}

// .
func (bfi *BFInterp) Put() {
	bfi.out.WriteByte(bfi.mem[bfi.ptr])
	bfi.pc += 1
}

// ,
func (bfi *BFInterp) Get() {
	c, _ := bfi.in.ReadByte()
	bfi.mem[bfi.ptr] = c
	bfi.pc += 1
}

func (bfi *BFInterp) Jump() {
	if bfi.mem[bfi.ptr] == 0 {
		bfi.pc = bfi.jumps[bfi.pc]
	}
	bfi.pc += 1
}

func (bfi *BFInterp) Loop() {
	if bfi.mem[bfi.ptr] != 0 {
		bfi.pc = bfi.loops[bfi.pc]
	}
	bfi.pc += 1
}

// > compressed
func (bfi *BFInterp) Move() {
	bfi.ptr += bfi.moves[bfi.pc]
	bfi.pc += bfi.moves[bfi.pc]
}

// < compressed
func (bfi *BFInterp) Nove() {
	bfi.ptr -= bfi.noves[bfi.pc]
	bfi.pc += bfi.noves[bfi.pc]
}

// + compressed
func (bfi *BFInterp) Math() {
	bfi.mem[bfi.ptr] = byte(int(bfi.mem[bfi.ptr]) + bfi.maths[bfi.pc])
	bfi.pc += bfi.maths[bfi.pc]
}

// - compressed
func (bfi *BFInterp) Nath() {
	bfi.mem[bfi.ptr] = byte(int(bfi.mem[bfi.ptr]) - bfi.naths[bfi.pc])
	bfi.pc += bfi.naths[bfi.pc]
}
