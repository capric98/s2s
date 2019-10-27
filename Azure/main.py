#!/usr/bin/env python3
# coding: utf-8
import argparse
import time
import sys
try:
    sys.stdout = open(sys.stdout.fileno(), mode='w', encoding='utf8', buffering=1)
except Exception as e:
    print(e)

import Cognitive
import subout


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="A sub generating tool based on Azure Cognitives.")
    parser.add_argument("filename", metavar="Filename",type=str, help="A wav file with 16kHz samplerate, single channel.")
    #parser.add_argument("-l","--log",action="store_false", dest="withlog",default=False,help="With a simple output which could be edited easily.")

    args = parser.parse_args()
    start_time = time.time()

    result = Cognitive.cognitive(args.filename)

    print()
    print("Translating and writting sub...")
    with open(args.filename+".ass", "w") as main, open(args.filename+"_log.ass", "w") as log:
        subout.use(result, main, log)
    print("--- %s seconds ---" % (time.time() - start_time))