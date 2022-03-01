#! /usr/bin/python3

import os
import sys
from subprocess import Popen

def extractFile(filename):
    if not filename.endswith(".lox"):
        return

    print("extracting "+filename, end="    ")
    sys.stdout.flush()

    program = sys.argv[1]
    expectFile = filename.split(".lox")[0] + ".expect"
    with open(expectFile, 'w') as f:
        p = Popen([program, filename], stdout=f, stderr=f)
        p.communicate()

    print("done!")

def extractDir(dirname):
    root, subdirs, files = next(os.walk(dirname))
    for f in files:
        extractFile(os.path.join(root, f))
    for d in subdirs:
        extractDir(os.path.join(root, d))

def main():
    root = sys.argv[2]
    if os.path.isdir(root):
        extractDir(root)
    else:
        extractFile(root)

main()
