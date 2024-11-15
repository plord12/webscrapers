#!/usr/bin/env python3
#
# selenium script to get aviva balance
#
# AVIVA_USER=user AVIVA_PWD=password ./aviva.py
#

import time
import re 
import os
import platform
import undetected_chromedriver as uc
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.by import By
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support.ui import Select

# Note: use_subprocess=False fails on mac, should be set true
# Note: driver is assumed to be x86-64 on linux, so we copy locally installed chromedriver
#
if platform.system() == "Linux":
    driver = uc.Chrome(headless=False,use_subprocess=False,driver_executable_path=os.environ['HOME']+"/.local/share/undetected_chromedriver/chromedriver_copy")
else:
    driver = uc.Chrome(headless=False,use_subprocess=True)

try:
    driver.get('https://www.avivamymoney.co.uk/Login')

    wait = WebDriverWait(driver, 10)
    element = wait.until(EC.element_to_be_clickable((By.CSS_SELECTOR,'#onetrust-accept-btn-handler')))
    element.click()

    wait.until(EC.element_to_be_clickable((By.CSS_SELECTOR,'a.btn-primary:nth-child(1)')))

    driver.find_element(By.CSS_SELECTOR,'#Username').send_keys(os.environ['AVIVAMYMONEY_USERNAME'])
    driver.find_element(By.CSS_SELECTOR,'#Password').send_keys(os.environ['AVIVAMYMONEY_PASSWORD'])
    driver.find_element(By.CSS_SELECTOR,'a.btn-primary:nth-child(1)').click()

    wait.until(EC.element_to_be_clickable((By.CSS_SELECTOR,'#FirstLetter')))

    Select(driver.find_element(By.CSS_SELECTOR,'#FirstLetter')).select_by_visible_text(os.environ['AVIVAMYMONEY_WORD'][int(driver.find_element(By.CSS_SELECTOR,'#FirstElement_Index').get_attribute('value'))-1])
    Select(driver.find_element(By.CSS_SELECTOR,'#SecondLetter')).select_by_visible_text(os.environ['AVIVAMYMONEY_WORD'][int(driver.find_element(By.CSS_SELECTOR,'#SecondElement_Index').get_attribute('value'))-1])
    Select(driver.find_element(By.CSS_SELECTOR,'#ThirdLetter')).select_by_visible_text(os.environ['AVIVAMYMONEY_WORD'][int(driver.find_element(By.CSS_SELECTOR,'#ThirdElement_Index').get_attribute('value'))-1])
    driver.find_element(By.CSS_SELECTOR,'[name=Next]').click()

    wait.until(EC.element_to_be_clickable((By.CSS_SELECTOR,'.vspace-reset.text-size-42')))

    print(driver.find_element(By.CSS_SELECTOR,'.vspace-reset.text-size-42').get_attribute('innerHTML').replace(',', '').replace('£', ''))

except:
    driver.save_screenshot('debug-avivamymoney-'+os.environ['AVIVAMYMONEY_USERNAME']+'.png')
