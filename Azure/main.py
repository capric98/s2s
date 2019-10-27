#!/usr/bin/env python3
#coding: utf-8
import Cognitive
import subout
import argparse

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="A sub generating tool based on Azure Cognitives.")
    parser.add_argument("filename", metavar="Filename",type=str, help="A wav file with 16kHz samplerate, single channel.")
    #parser.add_argument("-l","--log",action="store_false", dest="withlog",default=False,help="With a simple output which could be edited easily.")

    args = parser.parse_args()
    result = Cognitive.cognitive(args.filename)

    print()
    print("Translating and writting sub...")
    with open(args.filename+".ass", "w") as main, open(args.filename+"_log.ass", "w") as log:
        subout.use(result, main, log)