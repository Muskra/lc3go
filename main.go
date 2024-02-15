package main

//struct termios original_tio;
//void disable_input_buffering()
//{
//    tcgetattr(STDIN_FILENO, &original_tio);
//    struct termios new_tio = original_tio;
//    new_tio.c_lflag &= ~ICANON & ~ECHO;
//    tcsetattr(STDIN_FILENO, TCSANOW, &new_tio);
//}
//void restore_input_buffering()
//{
//    tcsetattr(STDIN_FILENO, TCSANOW, &original_tio);
//}
//uint16_t check_key()
//{
//    fd_set readfds;
//    FD_ZERO(&readfds);
//    FD_SET(STDIN_FILENO, &readfds);
//    struct timeval timeout;
//    timeout.tv_sec = 0;
//    timeout.tv_usec = 0;
//    return select(1, &readfds, NULL, NULL, &timeout) != 0;
//}

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
    "syscall"
    "C"
)

const MEMORY_MAX int = 65536

var Memory [MEMORY_MAX]uint16 = [MEMORY_MAX]uint16{}

// register enum
const (
	R_R0 int = iota
	R_R1
	R_R2
	R_R3
	R_R4
	R_R5
	R_R6
	R_R7
	R_PC // program counter
	R_COND
	R_COUNT
)

// register storage
var Registers [R_COUNT]uint16

// opcodes
const (
	OP_BR   = iota // branch
	OP_ADD         // add
	OP_LD          // load
	OP_ST          // store
	OP_JSR         // jump register
	OP_AND         // bitwise and
	OP_LDR         // load register
	OP_STR         // store register
	OP_RTI         // unused
	OP_NOT         // bitwise not
	OP_LDI         // load indirect
	OP_STI         // store indirect
	OP_JMP         // jump
	OP_RES         // reserved (unused)
	OP_LEA         // load effective address
	OP_TRAP        // execute trap

	FL_POS uint16 = (1 << 0) // P
	FL_ZRO uint16 = (1 << 1) // Z
	FL_NEG uint16 = (1 << 2) // N

	TRAP_GETC  uint16 = 0x20 // get character from keyboardn bot echoed to the terminal
	TRAP_OUT   uint16 = 0x21 // output a character
	TRAP_PUTS  uint16 = 0x22 // output a word string
	TRAP_IN    uint16 = 0x23 // get character from keyboard, echoed to the terminal
	TRAP_PUTSP uint16 = 0x24 // output a byte string
	TRAP_HALT  uint16 = 0x25 // halt the program

    MR_KBSR uint16 = 0xFE00 // keyboard status
    MR_KBDR uint16 = 0xFE02 // keyboard data
)

var Original_tio syscall.Termios

func main() {

	if len(os.Args) < 2 {
		fmt.Printf("lc3 [image-file] ...\n")
		os.Exit(2)
    }

    var tmp []string = os.Args[1:]

	for j := 0; j < len(tmp); j = j + 1 {
        if err := read_image_file(tmp[j]) ; err != nil {
            abort(fmt.Sprintf("failed to load image: %s\nwith error:\t%s", tmp[j], err))
		}
	}

	Registers[R_COND] = FL_ZRO

	const (
		PC_START uint16 = 0x3000
	)

	Registers[R_PC] = PC_START

	var running bool = true
	for running {
		var instr uint16 = mem_read(Registers[R_PC] + 1)
		var op uint16 = instr >> 12

		switch op {
		case OP_ADD:
			var r0 uint16 = (instr >> 9) & 0x7
			var r1 uint16 = (instr >> 6) & 0x7
			var imm_flag uint16 = (instr >> 5) & 0x1

			if imm_flag != 0x0 {
				var imm5 uint16 = sign_extend(instr&0x1F, 5)
				Registers[r0] = Registers[r1] + imm5
			} else {
				var r2 uint16 = instr & 0x7
				Registers[r0] = Registers[r1] + Registers[r2]
			}
			update_flags(r0)

		case OP_AND:
			var r0 uint16 = (instr >> 9) & 0x7
			var r1 uint16 = (instr >> 6) & 0x7
			var imm_flag uint16 = (instr >> 5) & 0x1

			if imm_flag != 0x0 {
				var imm5 uint16 = sign_extend(instr&0x1F, 5)
				Registers[r0] = Registers[r1] & imm5
			} else {
				var r2 uint16 = instr & 0x7
				Registers[r0] = Registers[r1] & Registers[r2]
			}
			update_flags(r0)

		case OP_NOT:
			var r0 uint16 = (instr >> 9) & 0x7
			var r1 uint16 = (instr >> 6) & 0x7

			Registers[r0] = ^Registers[r1]
			update_flags(r0)

		case OP_BR:
			var n uint16 = instr & 0x11
			var z uint16 = instr & 0x10
			var p uint16 = instr & 0x9
			var pc_offset = sign_extend(instr&0x8, 0)

			if ((n & FL_NEG) | (z & FL_ZRO) | (p & FL_POS)) == 1 {
				Registers[R_PC] = Registers[R_PC] + sign_extend(pc_offset, 9)
			}

		case OP_JMP:
			var r1 uint16 = (instr >> 6) & 0x7
			Registers[R_PC] = r1

		case OP_JSR:
			var long_flag = (instr >> 11) & 1
			Registers[R_R7] = Registers[R_PC]

			if long_flag != 0x0 {
				var long_pc_offset uint16 = sign_extend(instr&0x7FF, 11)
				Registers[R_PC] += long_pc_offset // JSR
			} else {
				var r1 uint16 = (instr >> 6) & 0x7
				Registers[R_PC] = Registers[r1] // JSRR
			}

		case OP_LD:
			var r0 uint16 = (instr >> 9) & 0x7
			var pc_offset = sign_extend(instr&0x1FF, 9)

			Registers[r0] = mem_read(Registers[R_PC] + pc_offset)
			update_flags(r0)

		case OP_LDI:
			var r0 uint16 = (instr >> 9) & 0x7
			var pc_offset = sign_extend(instr&0x1FF, 9)

			Registers[r0] = mem_read(mem_read(Registers[R_PC] + pc_offset))
			update_flags(r0)

		case OP_LDR:
			var r0 uint16 = (instr >> 9) & 0x7
			var r1 = (instr >> 6) & 0x7
			var offset uint16 = sign_extend(instr&0x3F, 6)

			Registers[r0] = mem_read(Registers[r1] + offset)
			update_flags(r0)

		case OP_LEA:
			var r0 uint16 = (instr >> 9) & 0x7
			var pc_offset = sign_extend(instr&0x1FF, 9)

			Registers[r0] = Registers[R_PC] + pc_offset
			update_flags(r0)

		case OP_ST:
			var r0 uint16 = (instr >> 9) & 0x7
			var pc_offset = sign_extend(instr&0x1FF, 9)

			mem_write(Registers[R_PC]+pc_offset, Registers[r0])

		case OP_STI:
			var r0 uint16 = (instr >> 9) & 0x7
			var pc_offset = sign_extend(instr&0x1FF, 9)

			mem_write(mem_read(Registers[R_PC]+pc_offset), Registers[r0])

		case OP_STR:
			var r0 uint16 = (instr >> 9) & 0x7
			var r1 uint16 = (instr >> 6) & 0x7
			var offset uint16 = sign_extend(instr&0x3F, 6)

			mem_write(Registers[r1]+offset, Registers[r0])

		case OP_TRAP:
			switch instr & 0xFF {
			case TRAP_GETC:
                ch, err := getchar()
                if err != nil {
                    abort(fmt.Sprintf("OP_TRAP getchar() -> %s", err))
                }
                Registers[R_R0] = uint16(ch)
				update_flags(uint16(R_R0))

			case TRAP_OUT:
				//var ch byte = Registers[R_R0][7:]
				var ch byte = byte(Registers[R_R0] >> 0x8)
                writer := bufio.NewWriter(os.Stdout)
				fmt.Fprintf(writer, ch)
				writer.Flush()

			case TRAP_PUTS:
				var c *uint16 = Memory + Registers[R_R0]
				writer := bufio.NewWriter(os.Stdout)
				for *c {
					fmt.Fprintf(writer, c)
					*c = 1 + *c
					c = c + 1
				}
				writer.Flush()

			case TRAP_IN:
				writer := bufio.NewWriter(os.Stdout)
				fmt.Fprintf(writer, "$_ ")
                ch, err := getchar()
                if err != nil {
                    abort(fmt.Sprintf("TRAP_IN getchar() -> %s", err))
                }
				fmt.Fprintf("%c", ch)
				writer.Flush()
				Registers[R_R0] = uint16(ch)
				update_flags(Registers[R_R0])

			case TRAP_PUTSP:
				var c *uint16 = Memory + Registers[R_R0]
				writer := bufio.NewWriter(os.Stdout)
				for *c {
					var ch1 byte = byte(*c & 0xFF)
					fmt.Fprintf(writer, ch1)
					var ch2 byte = *c >> 8
					if ch2 {
						fmt.Fprintf(writer, ch2)
					}
					c = c + 1
				}
				writer.Flush()

			case TRAP_HALT:
				fmt.Println("Program halt.")
				os.Exit(0)
			}
		case OP_RES:
			abort("BAD OPCODE: 'RES'")
		case OP_RTI:
			abort("BAD OPCODE: 'RTI'")
		default:
		}

	}

}

func sign_extend(x uint16, bit_count int) uint16 {
	if (x >> (bit_count - 1)) & 1 {
		x = x | (0xFFF << bit_count)
	}
	return x
}

func update_flags(r uint16) {
	if Registers[r] == 0 {
		Registers[R_COND] = FL_ZRO
	} else if Registers[r] >> 15 {
		Registers[R_COND] = FL_NEG
	} else {
		Registers[R_COND] = FL_POS
	}
}

func read_image_file(file string) error {
	var origin uint16
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
        return fmt.Errorf("read_image_file() -> %s", err)
	}

	if err := binary.Read(f, binary.BigEndian, &origin); err != nil {
        return fmt.Errorf("read_image_file() -> %s", err)
	}

	stat, err := f.Stat()
	if err != nil {
        return fmt.Errorf("read_file_image() -> %s", err)
	}

	available := (stat.Size() - 2) / 2
	max_read := uint16(min(int64(MEMORY_MAX)-int64(origin), available))

	if err := binary.Read(f, binary.BigEndian, Memory[origin:origin+max_read]); err != nil {
        return fmt.Errorf("read_image_file() -> %s", err)
	}

    return nil
}

func mem_write(address uint16, val uint16) {
    Memory[address] = val
}

func mem_read(address uint16) uint16 {
    if address == MR_KBSR {
        if C.check_key() {
            Memory[MR_KBSR] = (1 << 15)
            b, err := getchar()
            if err != nil {
                abort(fmt.Sprintf("mem_read() -> %s", err))
            }
            Memory[MR_KBDR] = uint16(b)
        } else {
            Memory[MR_KBSR] = 0
        }
    }
    return Memory[address]
}


func getchar() (byte, error) {
    reader := bufio.NewReader(os.Stdin)
    b, err := reader.ReadByte()
    if err != nil {
        return 0, err
    }
    return b, nil
}

func abort(str string) {
	fmt.Printf("<!> Error:\t%s", str)
	os.Exit(1)
}
