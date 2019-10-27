#!/usr/bin/env python3
# coding: utf-8
import hashlib
import requests
import time
import uuid
import math

appKey = ""
appPass = ""

def encrypt(signStr):
    hash_algorithm = hashlib.sha256()
    hash_algorithm.update(signStr.encode('utf-8'))
    return hash_algorithm.hexdigest()

def truncate(q):
    if q is None:
        return None
    size = len(q)
    return q if size <= 20 else q[0:10] + str(size) + q[size - 10:size]


def translate(qstr):
    curtime = "{:d}".format(math.floor(time.time()))
    salt = str(uuid.uuid1())
    payload = {
        "q": qstr,
        "from": "ja",
        "to": "zh-CHS",
        "appKey": appKey,
        "salt": salt,
        "sign": encrypt(appKey+truncate(qstr)+salt+curtime+appPass),
        "signType": "v3",
        "curtime": curtime
    }
    resp = requests.post("https://openapi.youdao.com/api", data=payload, headers={'Content-Type': 'application/x-www-form-urlencoded'})
    return resp.json()["translation"][0]

def translate_list(sentences):
    result = []
    for s in sentences:
        result.append(translate(s))
    return result
