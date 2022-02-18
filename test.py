#! /usr/bin/python3

import os

def testFile(filename):
    expect = ""
    with open(filename, 'r') as f:
        expect = f.read()

    output = os.popen('./golox '+filename).read()
    print(expect)
    print(output)

def compare(expect, output):

def main():
    testFile("test/function/recursion.lox")

main()
