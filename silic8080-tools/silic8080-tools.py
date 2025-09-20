import colorama
import opcodes
import sys
import os

colorama.init(autoreset=True)

def little_endian_converter(value):
    if value <= 0xFF:
        return f"{value:02X}"
    else:
        low = value & 0xFF
        high = (value >> 8) & 0xFF
        return f"{low:02X} {high:02X}"

def asmtobinary(file):
    result = list()
    connectedpair = None
    connectedpairwithcomma = None
    for line in str(file).splitlines():
        parts = str(line).split(",")
        connectedpairwithcomma = line.replace(", ", ",")
        if "," not in list(line):
            parts = str(line).split(" ")
            try:
                connectedpair = parts[0] + " " + parts[1]
            except IndexError:
                connectedpair = parts
        if parts[0] in opcodes.opcodes.values():
            result.append(f"{list(opcodes.opcodes.keys())[list(opcodes.opcodes.values()).index(parts[0])]:02X}")
            if len(parts) > 1:
                try:
                    idk = little_endian_converter(int(parts[-1].replace("H", "").strip(), 16)).split()
                    result.extend(idk)
                except Exception as e:
                    print(colorama.Fore.RED + colorama.Style.BRIGHT + "error in line!\n" + colorama.Fore.RESET + line)
                    return 1
        else:
            try:
                if connectedpair in opcodes.opcodes.values():
                    result.append(f"{list(opcodes.opcodes.keys())[list(opcodes.opcodes.values()).index(connectedpair)]:02X}")
                else:
                    if connectedpairwithcomma in opcodes.opcodes.values():
                        result.append(f"{list(opcodes.opcodes.keys())[list(opcodes.opcodes.values()).index(connectedpairwithcomma)]:02X}")
                    else:
                        print(colorama.Fore.RED + colorama.Style.BRIGHT + "bad mnemonic(s)!\n" + colorama.Fore.RESET + line)
                        return 1
            except Exception as e:
                print(colorama.Fore.RED + colorama.Style.BRIGHT + "bad mnemonic(s)!\n" + colorama.Fore.RESET + line)
                return 1
        connectedpair = None
        connectedpairwithcomma = None
    try:
        binary = bytes(int(b, 16) for b in result)
        with open(os.path.splitext(sys.argv[2])[0] + ".bin", "wb") as binfile:
            binfile.write(binary)
        print(colorama.Fore.GREEN + colorama.Style.BRIGHT + "saved to " + os.path.splitext(sys.argv[2])[0] + ".bin!")
    except Exception as e:
        print(colorama.Fore.RED + colorama.Style.BRIGHT + "binary conversion error!\n" + colorama.Fore.RESET + str(e))

if len(sys.argv) == 1:
    print(colorama.Fore.RED + colorama.Style.BRIGHT + "no args found!")
    exit()

if sys.argv[1] == "asmtobin":
    if len(sys.argv) == 2:
        print(colorama.Fore.RED + colorama.Style.BRIGHT + "not enough args! usage: asmtobin [filename.asm] ")
    else:
        with open(sys.argv[2], "r") as file:
            asmtobinary(file.read())
else:
    print(colorama.Fore.RED + colorama.Style.BRIGHT + "\"" + sys.argv[1] + "\" is not found!")
