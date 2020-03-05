import docker
import os
import time
from random import choice
from pprint import pprint

root_backup_folder = '/backups'
invalid_mounts = ['/var/run', '/dev']


class bcolors:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'

def print_stat(container_name, filename):
    random_char = ['_', '\\', '/', '|', '-', '.']
    print(' [{random}] Dumping contents of {container}: {filename}'.format(container=container.name[: container.name.index('.') if '.' in container.name else len(container.name)], filename=filename, random=choice(random_char)), end='\r')

def dump_data(container, file_or_folder, destination_path):
    f = open('{}.tar'.format(destination_path), 'wb')
    bits, stat = container.get_archive(file_or_folder)

    for chunk in bits:
        print_stat(container.name, file_or_folder)
        f.write(chunk)
    f.close()

def is_valid_mount(mount):
    for invalid_mount in invalid_mounts:
        if invalid_mount in mount:
            return False
    
    return True

client = docker.from_env()
containers = client.containers.list()

for container in containers:
    mounts = container.attrs['Mounts']
    container_name = container.name[: container.name.index('.') if '.' in container.name else len(container.name)]

    clean_mounts = list(filter(lambda mount: is_valid_mount(mount['Source']), mounts))
    destination_path = '{root_backup_folder}/{container}/{date}/'.format(root_backup_folder=root_backup_folder, container=container_name, date=time.strftime('%m-%d-%y'))

    os.makedirs(destination_path, exist_ok=True)
    for mount in clean_mounts:
        filename = mount['Destination'].replace('/', '-')[1:]        
        dump_data(container, mount['Destination'], destination_path + filename)
        
        print(' [{color_green}OK{end_color}] Dumping contents of {container}: {filename}'.format(container=container_name, filename=mount['Destination'], color_green=bcolors.OKGREEN, end_color=bcolors.ENDC))