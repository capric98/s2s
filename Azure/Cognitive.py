#!/usr/bin/env python3
#coding: utf-8
import sys
import time
try:
    import azure.cognitiveservices.speech as speechsdk
except ImportError:
    print("""
    Importing the Speech SDK for Python failed.
    Refer to
    https://docs.microsoft.com/azure/cognitive-services/speech-service/quickstart-python for
    installation instructions.
    """)
    sys.exit(1)

speech_key, service_region = "", ""

def cognitive(filename):
    """performs continuous speech recognition with input from an audio file"""
    # <SpeechContinuousRecognitionWithFile>
    speech_config = speechsdk.SpeechConfig(
        subscription=speech_key,
        region=service_region,
        speech_recognition_language="ja-JP"
    )
    audio_config = speechsdk.audio.AudioConfig(filename=filename)

    speech_recognizer = speechsdk.SpeechRecognizer(speech_config=speech_config, audio_config=audio_config)

    done = False
    log = []
    result = []

    def stop_cb(evt):
        """callback that stops continuous recognition upon receiving an event `evt`"""
        print('CLOSING on {}'.format(evt))
        speech_recognizer.stop_continuous_recognition()
        nonlocal done
        done = True
    def logevt(evt):
        nonlocal log
        evtlog = {
            "timestamp": (evt.result.offset+evt.result.duration)/10000000,
            "text"     : evt.result.text
        }
        log.append(evtlog)
    def checkpoint(evt):
        nonlocal log
        nonlocal result
        tmp = {
            "result": evt.result.text,
            "offset": evt.result.offset/10000000,
            "log"   : log
        }
        result.append(tmp)
        log = []
        print("Till:", (evt.result.offset+evt.result.duration)/10000000, evt.result.text)

    # Connect callbacks to the events fired by the speech recognizer
    speech_recognizer.recognizing.connect(logevt)
    speech_recognizer.recognized.connect(checkpoint)
    speech_recognizer.session_started.connect(lambda evt: print('SESSION STARTED: {}'.format(evt)))
    speech_recognizer.session_stopped.connect(lambda evt: print('SESSION STOPPED {}'.format(evt)))
    speech_recognizer.canceled.connect(lambda evt: print('CANCELED {}'.format(evt)))
    # stop continuous recognition on either session stopped or canceled events
    speech_recognizer.session_stopped.connect(stop_cb)
    speech_recognizer.canceled.connect(stop_cb)

    # Start continuous speech recognition
    speech_recognizer.start_continuous_recognition()
    while not done:
        time.sleep(.5)

    return result

if __name__ == "__main__":
    print(cognitive("gangan.wav"))