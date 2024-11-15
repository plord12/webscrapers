#!/usr/bin/env python3
#
# selenium script to get aviva balance
#
# AVIVA_USER=user AVIVA_PWD=password ./aviva.py
#

import time
import re 
import os 
import os.path
import sys
import platform
import undetected_chromedriver as uc
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.by import By
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait

# clean up
#
if 'OTP_CLEANCOMMAND' in os.environ:
    print("Running "+os.environ['OTP_CLEANCOMMAND'], file=sys.stderr)
    os.system(os.environ['OTP_CLEANCOMMAND'])
if os.path.exists('otp/aviva'):
  os.remove('otp/aviva')

# Note: use_subprocess=False fails on mac, should be set true
# Note: driver is assumed to be x86-64 on linux, so we copy locally installed chromedriver
#
if platform.system() == "Linux":
    driver = uc.Chrome(headless=False,use_subprocess=False,driver_executable_path=os.environ['HOME']+"/.local/share/undetected_chromedriver/chromedriver_copy")
else:
    driver = uc.Chrome(headless=False,use_subprocess=True)

try:
    driver.get('https://www.direct.aviva.co.uk/MyAccount/login')

    WebDriverWait(driver, 10).until(EC.element_to_be_clickable((By.CSS_SELECTOR,'#onetrust-accept-btn-handler'))).click()

    WebDriverWait(driver, 10).until(EC.element_to_be_clickable((By.CSS_SELECTOR,'#loginButton')))
    driver.find_element(By.CSS_SELECTOR,'#username').send_keys(os.environ['AVIVA_USERNAME'])
    driver.find_element(By.CSS_SELECTOR,'#password').send_keys(os.environ['AVIVA_PASSWORD'])
    driver.find_element(By.CSS_SELECTOR,'#loginButton').click()

    time.sleep(5)

    # otp
    count = 100
    while not os.path.isfile('otp/aviva') and count > 0:
        if 'OTP_COMMAND' in os.environ:
            print("Running "+os.environ['OTP_COMMAND'], file=sys.stderr)
            os.system(os.environ['OTP_COMMAND'])
        time.sleep(1)
        count = count-1
    f = open('otp/aviva', 'r')
    otp = re.findall("code is: ([0-9]*).", f.read())[0]
    f.close()
    print("otp="+otp, file=sys.stderr)
    os.remove('otp/aviva')

    driver.find_element(By.CSS_SELECTOR,'#factor').send_keys(otp)
    driver.find_element(By.CSS_SELECTOR,'#VerifyMFA').click()
    time.sleep(5)

    driver.find_element(By.PARTIAL_LINK_TEXT, 'Details').click()
    time.sleep(20)

    x = re.search('\"yourPensionValue\"[^0-9]*([0-9,.]*)[^0-9]*', driver.page_source) 

    print(x[1].replace(',', '')) 

except:
    driver.save_screenshot('debug-aviva-'+os.environ['AVIVA_USERNAME']+'.png')
