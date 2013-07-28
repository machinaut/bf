package bf

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
)

const (
	NEXT  = byte('>')
	PREV  = byte('<')
	INC   = byte('+')
	DEC   = byte('-')
	PUT   = byte('.')
	GET   = byte(',')
	JUMP  = byte('[')
	LOOP  = byte(']')
	CNEXT = byte('{')
	CPREV = byte('}')
	CINC  = byte('&')
	CDEC  = byte('|')
	NOP   = byte(0)
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
}

func NewBFInterp(prog io.Reader, in io.Reader, out io.Writer) *BFInterp {
	bfi := new(BFInterp)
	bfi.prog, _ = ioutil.ReadAll(prog)
	bfi.in = bufio.NewReader(in)
	bfi.out = bufio.NewWriter(out)
	bfi.jumps = make(map[int]int)
	bfi.loops = make(map[int]int)

	return bfi
}

// Run the program
func (bfi *BFInterp) Run() {
	// preprocess the program to compress instructions
	bfi.comp()
	// re-pack the program (remove NOPs)
	bfi.pack()
	// scan the program for jumps/loops
	bfi.scan()

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
	case CNEXT:
		bfi.CNext()
	case CPREV:
		bfi.CPrev()
	case CINC:
		bfi.CInc()
	case CDEC:
		bfi.CDec()
	default:
		//fmt.Println("ERROR: Bad Instruction: PC %d Op %d", bfi.pc, cmd)
		// continue anyways, ignore error
		bfi.pc += 1
	}
}

// Compress repeated instructions
func (bfi *BFInterp) comp() {
	for idx, cmd := range bfi.prog {
		// TODO: clean this up.  This shit is nasty.
		if cmd == INC || cmd == DEC || cmd == NEXT || cmd == NEXT {
			// scan forward to find length of the 'run'
			var i int
			for i = 0; i+idx < len(bfi.prog) && bfi.prog[i+idx] == cmd; i++ {
				// TODO: theres probably a better way to do this :P
			}
			// TODO: handle the overflow case
			if (cmd == NEXT || cmd == PREV) && i > 255 {
				fmt.Println("ERROR: run of CMD: %c too long to compress: %d", cmd, i)
				// continue anyways
			}
			// Compress if we find a run of many
			if i > 1 {
				// Overwrite command
				switch cmd {
				case NEXT:
					bfi.prog[idx] = CNEXT
				case PREV:
					bfi.prog[idx] = CPREV
				case INC:
					bfi.prog[idx] = CINC
				case DEC:
					bfi.prog[idx] = CDEC
				}
				// Byte following compressed command is number to apply
				bfi.prog[idx+1] = byte(i % 256)
				// Overwrite rest of the commands with no-ops
				for j := 2; j < i; j++ {
					bfi.prog[idx+j] = NOP
				}
			}
		}
	}
}

// Re-pack instruction stream to ignore nops
func (bfi *BFInterp) pack() {
	p := make([]byte, len(bfi.prog))
	j := 0 // index into p
	for i := 0; i < len(bfi.prog); i++ {
		cmd := bfi.prog[i]
		switch cmd {
		case NEXT, PREV, INC, DEC, PUT, GET, JUMP, LOOP:
			p[j] = cmd
			j++
		case CNEXT, CPREV, CINC, CDEC:
			p[j] = cmd
			p[j+1] = bfi.prog[i+1]
			j += 2
			i++ // skip next byte
		}
	}
	// Replace old program with new
	bfi.prog = p[0:j]
}

// Scan and find matching brace sets
func (bfi *BFInterp) scan() {
	// quick-n-dirty stack made out of a slice
	var stack []int

	for idx, cmd := range bfi.prog {
		// Memoize jumps
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
	fmt.Println("Jumps:", bfi.jumps)
	fmt.Println("Loops:", bfi.loops)
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
func (bfi *BFInterp) CNext() {
	bfi.ptr += int(bfi.prog[bfi.pc+1])
	bfi.pc += 2
}

// < compressed
func (bfi *BFInterp) CPrev() {
	bfi.ptr -= int(bfi.prog[bfi.pc+1])
	bfi.pc += 2
}

// + compressed
func (bfi *BFInterp) CInc() {
	//fmt.Printf("CInc ")
	//fmt.Printf("pc:%d ", bfi.pc)
	//fmt.Printf("cmd:%d ", bfi.prog[bfi.pc])
	//fmt.Printf("val:%d ", bfi.prog[bfi.pc+1])
	//fmt.Printf("ptr:%d ", bfi.ptr)
	//fmt.Printf("mem:%d", bfi.mem[bfi.ptr])
	//fmt.Printf("\n")
	bfi.mem[bfi.ptr] += bfi.prog[bfi.pc+1]
	bfi.pc += 2
}

// - compressed
func (bfi *BFInterp) CDec() {
	bfi.mem[bfi.ptr] -= bfi.prog[bfi.pc+1]
	bfi.pc += 2
}
