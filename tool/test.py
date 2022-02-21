#! /usr/bin/python3

import os
import sys
import time

ExitOnError = False

Total = 0
Failed = []

def report():
    global Total, Failed
    passed = Total-len(Failed)
    print("=== Total: %d Passed: %d Pass Rate: %.2f%%" %(Total, passed, passed/Total*100))
    print("--- Failed:")
    for f in Failed:
        print("--- "+f)

def analyzeOutput(output):
    result = ""
    it = iter(output.splitlines(True))
    for line in it:
        # print(line, end="")
        if line.startswith("error: "):
            next(it)
            next(it)
        else:
            result += line
    return result

def testFile(filename):
    global Total, Failed
    Total += 1

    start = time.time()

    notation = "// expect: "
    expect = ""

    lines = open(filename, 'r').read().splitlines(True)
    for line in lines:
        expect += line.partition(notation)[2]

    output = os.popen('./golox '+filename).read()
    result = analyzeOutput(output)

    elapsed = time.time() - start

    if expect != result:
        Failed.append(filename)
        print("=== FAIL: %s (%0.2f)s" %(filename, elapsed))
        print("--- Get:")
        print(result, end="")
        print("--- Expect:")
        print(expect, end="")
        if ExitOnError:
            sys.exit(1)
    else:
        print("=== PASS: %s (%0.2f)s" %(filename, elapsed))

def testDir(dirname):
    root, subdirs, files = next(os.walk(dirname))
    for f in files:
        testFile(os.path.join(root, f))
    for d in subdirs:
        testDir(os.path.join(root, d))

def main():
    walkDir = sys.argv[1]
    testDir(walkDir)
    report()

main()
