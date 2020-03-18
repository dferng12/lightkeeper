from random import choice
import datetime
import logging
import configparser

### Config
config = configparser.ConfigParser()
config.read('config.ini')
logFile = config['Logger']['LogDestination']

if logFile:
    logging.basicConfig(filename=logFile, filemode='a+', level=logging.INFO, format='%(asctime)s %(message)s', datefmt=config['Logger']['LogDateFormat'])

class bcolors:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'

### Functions
def parse_container_name(container_name):
    return container_name[: container_name.index('.') if '.' in container_name else len(container_name)]

def print_save_stat(container_name, filename, already_printed):
    random_char = ['_', '\\', '/', '|', '-', '.']
    if not logFile:
        print(f' [{choice(random_char)}] Dumping contents of {container_name}: {filename}', end='\r')
    else:
        if not already_printed:
            logging.info(f' [...] Dumping contents of {container_name}: {filename}')

def print_save_finished(container_name, filename):
    if not logFile:
        print(f' [{bcolors.OKGREEN}OK{bcolors.ENDC}] Dumping contents of {container_name}: {filename}')
    else:
        logging.info(f' [OK] Dumping contents of {container_name}: {filename}')

def print_info_deleted(folder, finished):
    status = f'{bcolors.OKGREEN}OK{bcolors.ENDC}' if finished else '...'
    if not logFile:
        print(f' [{status}] Deleting folder {folder}', end=('\n' if finished == 'OK' else '\r'))
    else:
        logging.info(f' [{status}] Deleting folder {folder}')
