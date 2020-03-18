import docker
import os
import time
import ast
import configparser
from logger import print_save_finished, print_save_stat, parse_container_name

### Config 
config = configparser.ConfigParser()
config.read('config.ini')

root_backup_folder = config['Public']['BackupFolder']
ignored_mounts = ast.literal_eval((config['Saver']['IgnoredMounts']))
quiet = config.getboolean('Public','Quiet')
date_format = config['Public']['DateFormat']

### Functions
def dump_data(container, file_or_folder, destination_path):
    f = open('{}.tar'.format(destination_path), 'wb')
    bits, _ = container.get_archive(file_or_folder)

    if not quiet: 
        container_name = parse_container_name(container.name)
        print_save_stat(container_name, file_or_folder, already_printed=False)

    for chunk in bits:
        if not quiet: print_save_stat(container_name, file_or_folder, already_printed=True)
        f.write(chunk)
    f.close()

    if not quiet: print_save_finished(container_name, destination_path)

def is_valid_mount(mount):
    for invalid_mount in ignored_mounts:
        if invalid_mount in mount:
            return False
    return True

def create_backup():
    client = docker.from_env()
    containers = client.containers.list()

    for container in containers:
        mounts = container.attrs['Mounts']
        container_name = parse_container_name(container.name)
        
        clean_mounts = list(filter(lambda mount: is_valid_mount(mount['Source']), mounts))
        destination_path = '{root_backup_folder}/{container}/{date}/'.format(root_backup_folder=root_backup_folder, container=container_name, date=time.strftime(date_format))

        os.makedirs(destination_path, exist_ok=True)
        for mount in clean_mounts:
            filename = mount['Destination'].replace('/', '-')[1:]        
            dump_data(container, mount['Destination'], destination_path + filename)

create_backup()