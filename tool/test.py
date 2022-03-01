#! /usr/bin/python3

import os
import sys
import time
import subprocess

ExitOnError = False

Total = 0
Failed = []

def report():
    global Total, Failed
    passed = Total-len(Failed)
    print("=== Total: %d Passed: %d Pass Rate: %.2f%%" %(Total, passed, passed/Total*100))
    if len(Failed) > 0:
        print("--- Failed:")
        for f in Failed:
            print("--- "+f)

def testFile(filename):
    if not filename.endswith(".lox"):
        return

    global Total, Failed
    Total += 1

    start = time.time()

    program = sys.argv[1]
    p = subprocess.Popen([program, filename], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    stdout, _ = p.communicate()
    result = stdout.decode("utf-8")

    expectFile = filename.split(".lox")[0] + ".expect"
    f = open(expectFile, 'r')
    expect = f.read()
    f.close()

    elapsed = time.time() - start

    if result != expect:
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
    root = sys.argv[2]
    if os.path.isdir(root):
        testDir(root)
    else:
        testFile(root)
    report()

main()
