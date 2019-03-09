import os 
import json
from collections import OrderedDict

DB_FILE = "DBSettings.json"
PUSH_FILE = "PushSettings.json"

def get_push_settings():
    push_dict = OrderedDict()
    client_id = input("Enter Pushbullet client id: ")
    push_dict["client_id"] = client_id 
    client_secret = input("Enter Pushbullet client secret: ")
    push_dict["client_secret"] = client_secret
    redirect_uri = input("Enter Pushbullet redirect URI: ")
    push_dict["redirect_uri"] = redirect_uri
    with open(PUSH_FILE, 'w') as push_file:
        push_file.write(json.dumps(push_dict, indent=4))

def get_db_settings():
    db_dict = OrderedDict()
    user = input("Enter database username: ")
    db_dict["username"] = user
    password = input("Enter database password: ")
    db_dict["password"] = password 
    db_name = input("Please enter database name: ")
    db_dict["db"] = db_name
    print(json.dumps(db_dict, indent=4))
    with open(DB_FILE, 'w') as db_file:
        db_file.write(json.dumps(db_dict, indent=4))

def get_settings():
    if os.path.exists(DB_FILE):
        os.remove(DB_FILE)
    if os.path.exists(PUSH_FILE):
        os.remove(PUSH_FILE)
    get_db_settings()
    get_push_settings()
    

def main():
    if os.path.exists(DB_FILE) or os.path.exists(PUSH_FILE):
        allowed = ['yes', 'no', 'y', 'n']
        res = input("Do you want to override old files? ").lower()
        while res not in allowed:
            print("Please enter one of the following: {}".format(", ".join(allowed)))
            res = input("Do you want to override old files? ").lower()
        if res in ['yes', 'y']:
            get_settings()
    else:
        get_settings()

if __name__ == '__main__':
    main()