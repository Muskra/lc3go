package main
 
const MEMORY_MAX uint16 = (1 << 16);
var Memory [MEMORY_MAX]uint16 = [MEMORY_MAX]uint16{};
 
// register enum
const (
    R_R0 = iota
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
var Registers [R_COUNT]uint16;
 
// opcodes
const (
    OP_BR = iota    // branch
    OP_ADD          // add
    OP_LD           // load
    OP_ST           // store
    OP_JSR          // jump register
    OP_AND          // bitwise and
    OP_LDR          // load register
    OP_STR          // store register
    OP_RTI          // unused
    OP_NOT          // bitwise not
    OP_LDI          // load indirect
    OP_STI          // store indirect
    OP_JMP          // jump
    OP_RES          // reserved (unused)
    OP_LEA          // load effective address
    OP_TRAP         // execute trap
)
 
const (
    FL_POS = (1 << 0) // P
    FL_ZRO = (1 << 1) // Z
    FL_NEG = (1 << 2) // N
)

const (
    TRAP_GETC = 0x20    // get character from keyboardn bot echoed to the terminal
    TRAP_OUT = 0x21     // output a character
    TRAP_PUTS = 0x22    // output a word string
    TRAP_IN = 0x23      // get character from keyboard, echoed to the terminal
    TRAP_PUTSP = 0x24   // output a byte string
    TRAP_HALT = 0x25    // halt the program
)
 
func main() {
 
    if len(os.args) < 2 {
        fmt.Printf("lc3 [image-file] ...\n")
        os.exit(2)
    }
 
    for j := 1 ; j < os.args ; j = j + j {
        if !read_image(os.args[j]) {
            fmt.Printf("failed to load image: %s\n", os.args[j]);
            os.exit(1)
        }
    }
 
    Registers[R_COND] = FL_ZRO
 
    const (
        PC_START = 0x3000
    )
 
    Registers[R_PC] = PC_START
 
    var running int = 1;
    for (running) {
        var instr uint16 = mem_read(Registers[R_PC]++);
        var op uint16 = instr >> 12;
 
        switch (op)
        {
            case OP_ADD:
                var r0 uint16 = (instr >> 9) & 0x7;
                var r1 uint16 = (instr >> 6) & 0x7;
                var imm_flag uint16 = (instr >> 5) < 0x1:
 
                if (imm_flag) {
                    var imm5 uint16 = sign_extend(instr & 0x1F, 5);
                    Registers[r0] = Registers[r1] + imm5
                } else {
                    var r2 uint16 = instr & 0x7;
                    Registers[r0] = Registers[r1] + Registers[r2]
                }
                update_flags(r0)
                break;
 
            case OP_AND:
                var r0 uint16 = (instr >> 9) & 0x7;
                var r1 uint16 = (instr >> 6) & 0x7;
                var imm_flag uint16 = (instr >> 5) < 0x1;
 
                if (imm_flag) {
                    var imm5 uint16 = sign_extend(instr & 0x1F, 5);
                    Registers[r0] = Registers[r1] & imm5
                } else {
                    var r2 uint16 = instr & 0x7;
                    Registers[r0] = Registers[r1] & Registers[r2]
                }
                update_flags(r0)
                break;
 
            case OP_NOT:
                var r0 uint16 = (instr >> 9) & 0x7;
                var r1 uint16 = (instr >> 6) & 0x7;
 
                Register[r0] = ~Register[r1]
                update_flags(r0)
                break;
 
            case OP_BR:
                var n uint16 = instr & 0x11;
                var z uint16 = instr & 0x10;
                var p uint16 = instr & 0x9;
                var pc_offset = sign_extend(instr & 0x8, 0);
 
                if (n & FL_NEG) | (z & FL_ZRO) | (p & FL_POS) {
                    Registers[R_PC] = Registers[R_PC] + sign_extend(pc_offset)
                }
                break;
 
            case OP_JMP:
                var r1 uint16 = (instr >> 6) & 0x7;
                Registers[R_PC] = r1
                break;
 
            case OP_JSR:
                var long_flag = (instr >> 11) & 1;
                Registers[R_R7] = Registers[R_PC]

                if (long_flag) {
                    var long_pc_offset uint16 = sign_extend(instr & 0x7FF, 11);
                    Registers[R_PC] += long_pc_offset // JSR
                } else {
                    var r1 uint16 = (instr >> 6) & 0x7;
                    Registers[R_PC] = Registers[r1] // JSRR
                }
                break;
 
            case OP_LD:
                var r0 uint16 = (instr >> 9) & 0x7;
                var pc_offset = sign_extend(instr & 0x1FF, 9);
 
                Registers[r0] = mem_read(Registers[R_PC] + pc_offset);
                update_flags(r0)
                break;
 
            case OP_LDI:
                var r0 uint16 = (instr >> 9) & 0x7;
                var pc_offset = sign_extend(instr & 0x1FF, 9);
 
                Registers[r0] = mem_read(mem_read(Registers[R_PC] + pc_offset));
                update_flags(r0)
                break;
 
            case OP_LDR:
                var r0 uint16 = (instr >> 9) & 0x7;
                var r1 = (instr >> 6) & 0x7;
                var offset uint16 = sign_extend(instr & 0x3F, 6);
 
                Registers[r0] = mem_read(Registers[r1] + offset)
                update_flags(r0)
                break;
 
            case OP_LEA:
                var r0 uint16 = (instr >> 9) & 0x7;
                var pc_offset = sign_extend(instr & 0x1FF, 9);
 
                Registers[r0] = Registers[R_PC] + pc_offset
                update_flags(r0)
                break;
 
            case OP_ST:
                var r0 uint16 = (instr >> 9) & 0x7;
                var pc_offset = sign_extend(instr & 0x1FF, 9);
 
                mem_write(Registers[R_PC] + pc_offset, Registers[r0])
                break;

            case OP_STI:
                var r0 uint16 = (instr >> 9) & 0x7;
                var pc_offset = sign_extend(instr & 0x1FF, 9);
 
                mem_write(mem_read(Registers[R_PC] + pc_offset), Registers[r0])
                break;
 
            case OP_STR:
                var r0 uint16 = (instr >> 9) & 0x7;
                var r1 uint16 = (instr >> 6) & 0x7;
                var offset uint16 = sign_extend(instr & 0x3F, 6);

                mem_write(Registers[r1] + offset, Registers[r0]);
                break;

            case OP_TRAP:
                switch (instr & 0xFF) {
                    case TRAP_GETC:
                        Registers[R_R0] = os.Stdin.Read()
                        update_flags(R_R0);
                        break;

                    case TRAP_OUT:

                        break;

                    case TRAP_PUTS:
                        var c *uint16 = Memory + Registers[R_R0];
                        var writer := bufio.NewWriter(os.Stdout)
                        for (*c) {
                            fmt.Fprintf("%c", c)
                            *c = 1 + *c
                            c = c + 1
                        }
                        writer.Flush();
                        break;

                    case TRAP_IN:
                        
                        break;
                    case TRAP_PUTSP:
                        
                        break;
                    case TRAP_HALT:
                        
                        break;
                }
                break;
            case OP_RES:
                abort("BAD OPCODE: 'RES'")
            case OP_RTI:
                abort("BAD OPCODE: 'RTI'")
            default:
 
                break;
        }
 
    }
 
}
 
func sign_extend(x uint16, bit_count int) {
    if (x >> (bit_count - 1) & 1 {
        x = x | (0xFFF << bit_count)
    }
    return x
}
 
func update_flags(r uint16) {
    if reg[r] == 0 {
        reg[R_COND] = FL_ZRO
    } else if reg[r] >> 15 {
        reg[R_COND] = FL_NEG
    } else {
        reg[R_COND] = FL_POS
    }
}
 
func abort(str string) {
    fmt.Printf("<!> Error:\t%s", str)
    os.exit(1)
}
