#!/usr/bin/python
#
# temp_controller.py

import time
import RPi.GPIO as GPIO
import temperature as temp
import urllib2

TAP1_INPUT_PIN = 22

FRIDGE_OUTPUT_PIN = 16
HIGH_TEMP = 42.0
LOW_TEMP = 38.0

class TapInputPin(object):
#  TIME_TO_END_POUR_IN_SEC = 1

  def __init__(self, pin_num):
    self.pin_num = pin_num
    GPIO.setup(pin_num, GPIO.IN, pull_up_down=GPIO.PUD_UP)
    GPIO.add_event_detect(pin_num, GPIO.FALLING, callback=self.tick, bouncetime=3)

  def tick(self):
    print "TICK!"

class OutputPin(object):
  def __init__(self, pin_num):
    self.pin_num = pin_num
    self.state = False
    GPIO.setup(pin_num, GPIO.OUT)

  def set(self, state):
    GPIO.output(self.pin_num, state)
    self.state = state

  def __str__(self):
    if self.state == True:
      return "on"
    else:
      return "off"

def read_temperature():
  '''Reads temp from sensor. Returns fahrenheit.'''
  with open("/sys/bus/w1/devices/28-000007605f5e/w1_slave") as sensorfile:
    sensor_text = sensorfile.read()
  sensor_data = sensor_text.split("\n")[1].split(" ")[9]
  temp = float(sensor_data[2:])
  temp = temp / 1000

  # Convert to fahrenheit
  return (temp * 1.8) + 32

if __name__ == '__main__':
  '''Run temperature regulation loop'''
  # Set up GPIO pin as output
  GPIO.setmode(GPIO.BOARD)
  fridge_pin = OutputPin(FRIDGE_OUTPUT_PIN)
  tap_pin = TapInputPin(TAP1_INPUT_PIN)
  try:
    # Run update loop
    fridge_state = False
    while True:
      temp = read_temperature()
      if temp > HIGH_TEMP:
        fridge_state = True
        fridge_pin.set(True)
      elif temp < LOW_TEMP:
        fridge_state = False
        fridge_pin.set(False)

      # Datalog
      try:
        urllib2.urlopen(
            "http://foo.com/store?secret=beerisgood&"
            "temp_f=%0.1f&keg1=0&keg2=0&fridge_on=%d"
            % (temp, fridge_state))
      except:
        pass

      # Sleep for 5s
      time.sleep(5.0)
  finally:
    # Release pin driver
    GPIO.cleanup()
