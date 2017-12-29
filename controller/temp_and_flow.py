import os
import re
import subprocess
import time
import urllib2
import time

import RPi.GPIO as GPIO

RELAY = 17
FLOW1 = 18
FLOW2 = 23

INIT = 18927
TURN_ON_TEMP = 50
TURN_OFF_TEMP = 40

URL = ('http://si-lax-beer-fridge.appspot.com/store?secret=beerisgood&'
       'temp_f=%0.2f&keg1=%d&keg2=%d&fridge_on=%d')

global count1, count2, compressor_on 
count1, count2, compressor_on = 0, 0, 0

def count_pulse1(_):
  global count1
  count1 += 1

def count_pulse2(_):
  global count2
  count2 += 1

def init():
  subprocess.check_call(['modprobe', 'w1-gpio'])
  subprocess.check_call(['modprobe', 'w1-therm'])
  GPIO.setmode(GPIO.BCM)
  GPIO.setwarnings(False)
  GPIO.setup(RELAY, GPIO.OUT)
  set_compressor_gpio(0)
  GPIO.setup(FLOW1, GPIO.IN, pull_up_down=GPIO.PUD_UP)
  GPIO.setup(FLOW2, GPIO.IN, pull_up_down=GPIO.PUD_UP)
  GPIO.add_event_detect(FLOW1, GPIO.FALLING, callback=count_pulse1)
  GPIO.add_event_detect(FLOW2, GPIO.FALLING, callback=count_pulse2)

def set_pin(i, high):
  print "Setting pin #%d to %d" % (i, int(high))
  val = GPIO.HIGH if high else GPIO.LOW
  GPIO.output(i, val)

def set_compressor_gpio(on):
  global compressor_on
  compressor_on = on
  set_pin(RELAY, on)

def drive_compressor(temp):
  if temp > TURN_ON_TEMP:
    set_compressor_gpio(True)
  if temp < TURN_OFF_TEMP:
    set_compressor_gpio(False)

def list_w1_devices():
  return [x for x in os.listdir('/sys/bus/w1/devices/')
          if not x.startswith('w1')]

def retrieve_temp(id):
  with open('/sys/bus/w1/devices/%s/w1_slave' % id) as f:
    return float(re.search(r't=(\d+)', f.read()).group(1))/1000.0

def fahrenheit(c):
  return 32.0 + 9.0/5.0*c

def keg_amounts():
  keg1 = INIT - .367*count1
  keg2 = INIT - .367*count2
  return keg1, keg2

if __name__ == '__main__':
  init()
  num_w1_ids = 0
  while num_w1_ids != 1:
    time.sleep(10)
    ids = list_w1_devices()
    print ids
    num_w1_ids = len(ids)

  try:
    while True:
      try:
        t = fahrenheit(retrieve_temp(ids[0]))
        drive_compressor(t)
      except e:
        print e

      keg1, keg2 = keg_amounts()
      print "Temp(F): %.2f; kegs(ml) = %d, %d; on = %d" % (t, keg1, keg2,
                                                           compressor_on)
      try:
        urllib2.urlopen(URL % (t, keg1, keg2, compressor_on))
      except urllib2.URLError as e:
        print e
      time.sleep(10)
  finally:
    print "Goodbye!"
    GPIO.cleanup()

