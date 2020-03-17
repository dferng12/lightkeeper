import os
import shutil
import time
import datetime
from pprint import pprint

root_backup_folder = './backups/' #WITH THE SLASH
history_amount = 4
date_format = '%m-%d-%Y'

def get_oldest_backups(backups):
    timestamps = {}
    for backup in backups:
        timestamps[backup] = time.mktime(datetime.datetime.strptime(backup, date_format).timetuple())

    return list(map(lambda item: item[0], sorted(timestamps.items(), key=lambda item: item[1])))[:-4]

def get_oldest_paths():
    services = []
    to_delete = []
    
    for r, d, _ in os.walk(root_backup_folder):
        if r == root_backup_folder:
            services = list(map(lambda item: root_backup_folder + item, d))

        if r in services:
            if len(d) > history_amount:
                oldest = get_oldest_backups(d)
                to_delete = [*list(map(lambda backup_name: f'{r}/{backup_name}', oldest)), *to_delete]
    
    return to_delete

def delete_folders(folders):
    for folder in folders:
        shutil.rmtree(folder)

delete_folders(get_oldest_paths())