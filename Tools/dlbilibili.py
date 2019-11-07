#!/usr/bin/env python3
# coding: utf-8
import argparse
import json
import requests
import re

downloader_path = "fetcher.exe"
ffmpeg_path = "C:\\ffmpeg\\bin\\ffmpeg.exe"

def download(info, aurl, filename):
    import os
    import subprocess
    import sys
    subprocess.run([downloader_path, aurl, info[0]["baseUrl"], info[1]["baseUrl"], filename], stdout=sys.stdout, stderr=sys.stderr)

    if info[1]["baseUrl"]=="":
        os.rename(filename+".v",filename+".flv")
    else:
        subprocess.run([ffmpeg_path, "-i", filename+".v", "-i", filename+".a", "-acodec", "copy", "-vcodec", "copy", filename+".mp4"],stdout=sys.stdout, stderr=sys.stderr)
        os.remove(filename+".v")
        os.remove(filename+".a")

def to_valid_name(str):
    invalid = '''\/:*?"<>| ;'''
    return "".join(c for c in str if c not in invalid)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="A bilibili download tool.")
    parser.add_argument("code", metavar="code",type=str, help="Bilibili video av code or ep code. (eg. av1234, ep123, av1234p5)")
    parser.add_argument("-c","--cookie", dest="cookie_file", default="", type=str, help="A cookie file, content copied from Chrome.")
    args = parser.parse_args()

    client = requests.Session()
    client.headers.update({"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36"})
    if args.cookie_file != "":
        with open(args.cookie_file, "r") as f:
            cookie = f.read()
            if cookie[-1]=="\n":
                cookie=cookie[:-1]
            client.headers.update({"Cookie": cookie})

    if args.code[:2]=="av":
        code = args.code[2:]
        pagelist = json.loads(client.get("https://api.bilibili.com/x/player/pagelist?aid="+code+"&jsonp=jsonp").text)
        if len(args.code.split("p"))!=1:
            p=int(args.code.split("p")[-1])-1
        else: p=0
        cid = str(pagelist["data"][p]["cid"])
        filename = json.loads(client.get("https://api.bilibili.com/x/web-interface/view?aid="+code+"&cid="+cid).text)
        filename = filename["data"]["title"]
        filename = filename+"_p"+str(p+1)

        url = "https://www.bilibili.com/video/"+args.code
        resp = client.get(url).text
    elif args.code[:2]=="ep":
        url = "https://www.bilibili.com/bangumi/play/"+args.code
        resp = client.get(url).text

        mwrapper = re.findall("media-wrapper.*?\</h1\>", resp)[0]
        filename = mwrapper.split(">")[2][:-4]
    else:
        print("Unknown:", code)

    info = re.findall('window\.__playinfo__=.*?}\</script\>', resp)[0]
    dic = json.loads(info[20:-9])
    try:
        dash = dic["data"]["dash"]
    except:
        dash = {
            "video": [
                {
                    "baseUrl": dic["data"]["durl"][0]["url"]
                }
            ],
            "audio": [
                {
                    "baseUrl": ""
                }
            ]
        }
    video_info = dash["video"]
    audio_info = dash["audio"]
    dinfo = [video_info[0], audio_info[0]]
    download(dinfo, url, to_valid_name(filename))
