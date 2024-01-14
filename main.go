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
                var imm_flag uint16 = (instr >> 5) < 0x1;

                if (imm_flag) {
                    var imm5 uint16 = sign_extend(instr & 0x1F, 5);
                    Registers[r0] = !(Registers[r1] | imm5)
                } else {
                    var r2 uint16 = instr & 0x7;
                    Registers[r0] = !(Registers[r1] | Registers[r2])
                break;
            case OP_BR:

                break;
            case OP_JMP:

                break;
            case OP_JSR:

                break;
            case OP_LD:

                break;
            case OP_LDI:
                var r0 uint16 = (instr >> 9) & 0x7;
                var pc_offset = sign_extend(instr & 0x1FF, 9);

                reg[r0] = mem_read(mem_read(reg[R_PC] + pc_offset));
                update_flags(r0)
                break;

            case OP_LDR:

                break;
            case OP_LEA:

                break;
            case OP_ST:

                break;
            case OP_STI:

                break;
            case OP_STR:

                break;
            case OP_TRAP:

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

func update_flags(r uint16) {
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
