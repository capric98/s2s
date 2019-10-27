#!/usr/bin/env python3
# coding: utf-8
from translate import translate_list

subhead = """[Script Info]
; Script generated by capric98/s2s
; https://github.com/capric98/s2s
Title: 879
ScriptType: v4.00+
PlayDepth: 0

[Aegisub Project Garbage]
Last Style Storage: Default
Active Line: 13

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default,微软雅黑,16,&H00FFFFFF,&H0300FFFF,&H00000000,&H02000000,0,0,0,0,100,100,0,0,3,1,1,1,10,10,10,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text"""

def cut(text):
    result = []
    while True:
        minp = 2147483647
        if text.find("。")!=-1:
            minp = min(minp, text.find("。"))
        if text.find("、")!=-1:
            minp = min(minp, text.find("、"))
        if text.find("，")!=-1:
            minp = min(minp, text.find("，"))
        if text.find("！")!=-1:
            minp = min(minp, text.find("！"))
        if text.find("？")!=-1:
            minp = min(minp, text.find("？"))
        if minp==2147483647:
            break

        print(text[:minp])
        result.append(text[:minp])
        text = text[minp+1:]

    if len(text)!=0:
        result.append(text)
    return result

def totimecode(timestamp):
    p = 1
    tc = [int(timestamp*100)%100,0,0,0]
    if timestamp>60*60*24:
        print("Duration too long!")
        return
    timestamp = int(timestamp)

    while timestamp>=60:
        tc[p] = int(timestamp)%60
        timestamp = int(timestamp/60)
        p+=1
    tc[p] = int(timestamp)

    return "{:02d}:{:02d}:{:02d}.{:02d}".format(tc[3], tc[2], tc[1], tc[0])

def arrangeline(text):
    result = ""
    while len(text)>23:
        result += text[:23]+"\\N"
        text = text[23:]
    result += text
    return result

def use(result, f0, f1):
    print(subhead, file=f0)
    print(subhead, file=f1)

    for section in result:
        sentences = cut(section["result"])
        if len(sentences)==0:
            continue
        trans = translate_list(sentences)
        print(trans)
        #print(section["log"])

        p = 0
        count = 0
        lastlog = section["offset"]
        for i in range(len(section["log"])):
            word = section["log"][i]
            nexttime = word["timestamp"]+0.5 if i+1==len(section["log"]) else section["log"][i+1]["timestamp"]
            substr = "Dialogue: 0,"+totimecode(word["timestamp"]-0.2)+","+totimecode(nexttime-0.2)+",Default,,0,0,0,,"
            
            if p!=0 and len(word["text"])-count<0:
                p-=1
                count-=len(sentences[p])
            if p>=len(section["log"]):
                p = len(section["log"])-1
            if len(word["text"])-count>len(sentences[p]):
                tmp = "Dialogue: 0,"+totimecode(lastlog)+","+totimecode(word["timestamp"])+",Default,,0,0,0,,"
                tmp += arrangeline(trans[p])+"\\N"
                tmp += arrangeline(sentences[p][:len(word["text"])-count])
                lastlog = word["timestamp"]
                print(tmp,file=f1)
                count+=len(sentences[p])
                p+=1
            substr+=arrangeline(trans[p])+"\\N"
            print(substr+arrangeline(sentences[p][:len(word["text"])-count]),file=f0)