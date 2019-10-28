#!/usr/bin/env python3
# coding: utf-8
import argparse
import json
import requests
import re

downloader_path = "fetcher.exe"
ffmpeg_path = "C:\\ffmpeg\\bin\\ffmpeg.exe"

def download(info, aurl, filename):
    import subprocess
    import sys
    subprocess.run([downloader_path, aurl, info[0]["baseUrl"], info[1]["baseUrl"], filename], stdout=sys.stdout, stderr=sys.stderr)

    if info[1]["baseUrl"]=="":
        import os
        os.rename(filename+".v",filename+".flv")
    else:
        subprocess.run([ffmpeg_path, "-i", filename+".v", "-i", filename+".a", "-acodec", "copy", "-vcodec", "copy", filename+".mp4"],stdout=sys.stdout, stderr=sys.stderr)
        os.remove(filename+".v")
        os.remove(filename+".a")
        


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="A bilibili download tool.")
    parser.add_argument("code", metavar="code",type=str, help="Bilibili video av code or ep code.")
    parser.add_argument("-c","--cookie", dest="cookie_file", default="", type=str, help="A cookie file, content copied from Chrome.")
    args = parser.parse_args()

    client = requests.Session()
    client.headers.update({"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36"})
    if args.cookie_file != "":
        with open(args.cookie_file, "r") as f:
            cookie = f.read()
            client.headers.update({"Cookie": cookie})

    if args.code[:2]=="av":
        code = code[2:]
        pagelist = json.loads(client.get("https://api.bilibili.com/x/player/pagelist?aid="+code+"&jsonp=jsonp").text)
        if args.url.split("/")[-1]!=code and args.url.split("/")[-1]!="":
            #print(args.url.split("/")[-1]!="")
            p=int(args.url.split("/")[-1][3:])-1
        else: p=0
        cid = str(pagelist["data"][p]["cid"])
        filename = json.loads(client.get("https://api.bilibili.com/x/web-interface/view?aid="+code+"&cid="+cid).text)
        filename = filename["data"]["title"]
        filename = filename+"_p"+str(p+1)

        url = "https://www.bilibili.com/video/"+args.code
        
        
    elif args.code[:2]=="ep":
        url = "https://www.bilibili.com/bangumi/play/"+args.code
    else:
        print("Unknown:", code)

    resp = client.get(url).text
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
    download(dinfo, url, "test")
    


# Get https://api.bilibili.com/x/player/pagelist?aid={VIDEO_AID}&jsonp=jsonp -> You can get cid here.
# {
#   "code": 0,
#   "message": "0",
#   "ttl": 1,
#   "data": [
#     {
#       "cid": 126014171,
#       "page": 1,
#       "from": "vupload",
#       "part": "[BE_6000k]TQ- What Does 'Audio Grade' Mean-",
#       "duration": 306,
#       "vid": "",
#       "weblink": "",
#       "dimension": {
#         "width": 1920,
#         "height": 1080,
#         "rotate": 0
#       }
#     }
#   ]
# }
# Get https://api.bilibili.com/x/web-interface/view?aid=73668103&cid=126014171
# {
#   "code": 0,
#   "message": "0",
#   "ttl": 1,
#   "data": {
#     "bvid": "BV1RE41187vG",
#     "aid": 73668103,
#     "videos": 1,
#     "tid": 191,
#     "tname": "影音智能",
#     "copyright": 1,
#     "pic": "http://i2.hdslb.com/bfs/archive/8f2d3fb8f5af2cb62795f21a23f6f495c62791fe.jpg",
#     "title": "【官方双语】“音频”电容是什么？ #电子速谈",
#     "pubdate": 1572196120,
#     "ctime": 1572196120,
#     "desc": "长期招募翻译nixiesubs.com/html/2015/recruit.html\n买衣服https://www.lttstore.com/\nLinus谈科技 https://space.bilibili.com/12434430 \nLMG论坛 https://linustechtips.com/",
#     "state": 0,
#     "attribute": 16512,
#     "duration": 306,
#     "rights": {
#       "bp": 0,
#       "elec": 0,
#       "download": 1,
#       "movie": 0,
#       "pay": 0,
#       "hd5": 0,
#       "no_reprint": 1,
#       "autoplay": 1,
#       "ugc_pay": 0,
#       "is_cooperation": 0,
#       "ugc_pay_preview": 0
#     },
#     "owner": {
#       "mid": 12564758,
#       "name": "TechQuickie",
#       "face": "http://i1.hdslb.com/bfs/face/66478178c0319d71db5c7338d0ced32ce64a6d16.jpg"
#     },
#     "stat": {
#       "aid": 73668103,
#       "view": 3437,
#       "danmaku": 16,
#       "reply": 35,
#       "favorite": 59,
#       "coin": 61,
#       "share": 16,
#       "now_rank": 0,
#       "his_rank": 0,
#       "like": 163,
#       "dislike": 0,
#       "evaluation": ""
#     },
#     "dynamic": "#TechQuickie##辉光字幕组##科普#",
#     "cid": 126014171,
#     "dimension": {
#       "width": 1920,
#       "height": 1080,
#       "rotate": 0
#     },
#     "no_cache": false,
#     "pages": [
#       {
#         "cid": 126014171,
#         "page": 1,
#         "from": "vupload",
#         "part": "[BE_6000k]TQ- What Does 'Audio Grade' Mean-",
#         "duration": 306,
#         "vid": "",
#         "weblink": "",
#         "dimension": {
#           "width": 1920,
#           "height": 1080,
#           "rotate": 0
#         }
#       }
#     ],
#     "subtitle": {
#       "allow_submit": false,
#       "list": []
#     }
#   }
# }

# https://upos-hz-mirrorwcsu.acgvideo.com/upgcxcode/17/33/87763317/87763317-1-30216.m4s?e=ig8euxZM2rNcNbdlhoNvNC8BqJIzNbfqXBvEqxTEto8BTrNvN0GvT90W5JZMkX_YN0MvXg8gNEV4NC8xNEV4N03eN0B5tZlqNxTEto8BTrNvNeZVuJ10Kj_g2UB02J0mN0B5tZlqNCNEto8BTrNvNC7MTX502C8f2jmMQJ6mqF2fka1mqx6gqj0eN0B599M=&uipk=5&nbs=1&deadline=1572216050&gen=playurl&os=wcsu&oi=3080740048&trid=6df2d61b6d754b348f6192699cfc92dbu&platform=pc&upsig=fb9d922a71794e5914c96c120d3faf66&uparams=e,uipk,nbs,deadline,gen,os,oi,trid,platform&mid=0
# https://upos-hz-mirrorks3u.acgvideo.com/upgcxcode/17/33/87763317/87763317-1-30112.m4s?e=ig8euxZM2rNcNbdlhoNvNC8BqJIzNbfqXBvEqxTEto8BTrNvN0GvT90W5JZMkX_YN0MvXg8gNEV4NC8xNEV4N03eN0B5tZlqNxTEto8BTrNvNeZVuJ10Kj_g2UB02J0mN0B5tZlqNCNEto8BTrNvNC7MTX502C8f2jmMQJ6mqF2fka1mqx6gqj0eN0B599M=&uipk=5&nbs=1&deadline=1572215542&gen=playurl&os=ks3u&oi=3080740048&trid=f4dc515ecc4240e9b15cab9a21cdc4e3u&platform=pc&upsig=cad19424cc4701f6710a50b06f80e30d&uparams=e,uipk,nbs,deadline,gen,os,oi,trid,platform&mid=3410879
# Origin        : https://www.bilibili.com
# Referer       : https://www.bilibili.com/video/av46965462/
# Sec-Fetch-Mode: cors
# Sec-Fetch-Site: cross-site
# User-Agent    : Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36
