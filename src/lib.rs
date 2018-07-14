#[macro_use]
extern crate diesel;
extern crate dotenv;
extern crate reqwest;
extern crate serde_json;
extern crate rocket_contrib;
#[macro_use]
extern crate serde_derive;

pub mod schema;
pub mod models;
pub mod searches_db;

pub mod subparser{

    //import necessary constructs for our module
    use reqwest::{get, Client};
    use serde_json::{Value, from_str};

    pub type SubResult = (String, String);
    
    /**
     * Get the first result for a search on a subreddit, sortng by new
     * @param sub - the subreddit to search in
     * @param search - the search query
     * @return url_map - a map of the comments link to the post title 
    */
    pub fn get_results(mut sub: String, mut search:String) 
    -> Result<SubResult, String>{

        //in case subreddit is missing r/
        if !sub.contains("r/"){
            sub = format!("r/{}", sub);
        }
        //prevent errors with link
        if search.contains(" "){
            search = search.replace(" ", "+");
        }

        let url = format!("https://www.reddit.com/{}/\
            search.json?q={}&sort=new&restrict_sr=on&limit=1", sub, search);

        //get the json text from the Reddit api
        let content = match get(&url).unwrap().text(){
            Ok(content) => content, 
            Err(_) => return Err("Error retrieving webpage".to_string())
        };

        //store this in a serde_json Value object 
        let json: Value = match from_str(&content){
            Ok(json) => json,
            Err(_) => return Err("Error decoding json object, \
            likely due to an invalid subreddit entered".to_string())
        };

        let results = json["data"]["children"].as_array().expect("Could not into array");

        //if there are no children, invalid sub
        if results.len() == 0{
            return Err("Invalid subreddit provided".to_string());
        }

        let result = results.get(0).unwrap();
        let perma = result["data"]["permalink"].as_str().unwrap();
        let link = format!("https://www.reddit.com{}", perma);
        let title = result["data"]["title"].as_str().unwrap();

        println!("{:#?}", (&link, &title));
        Ok((link, title.to_string()))
    }

    /**
     * Check whether or not a subreddit is valid 
     * @param sub - the subreddit to validate
     * @return - Result containing either a boolean or a string 
     * describing an error
    */
    pub fn validate_subreddit(mut sub: String) -> Result<bool, String> {

        //in case subreddit is missing r/
        if !sub.contains("r/"){
            sub = format!("r/{}", sub);
        }

        sub = sub.trim().to_string();

        //we use the about page because using post with reqwest wasn't working
        let url = format!("https://www.reddit.com/{}/\
        about.json", sub);

        let client = Client::new();

        let content = match client.get(&url).send().unwrap().text(){
            Ok(content) => {
                content
            }, 
            Err(e) => {
                //if we can't connect, return an error
                let error = format!("Error retrieving webpage: {}", e.to_string());
                return Err(error);
            }
        };

        let json: Value = from_str(&content).unwrap();

        //otherwise, check if there is a valid url for the subreddit in the json 
        //response 
        match json["data"]["url"].as_str() {
            Some(_) => Ok(true), 
            None => Ok(false)
        }
    }
}

pub mod pushbullet{

    use std::collections::HashMap;
    use reqwest::Client;
    use reqwest::header::{Headers, ContentType};
    use serde_json::{Value, from_str};
    use std::thread::Thread;
    use std::sync::mpsc;
    use std::sync::Mutex;
    use std::sync::mpsc::{Receiver, Sender};
    use super::subparser::{SubResult, get_results};
    use super::searches_db::searches_db::{get_interval, get_searches};

    //Constant urls for the Pushbullet APIs
    const DEVICES_URL: &str = "https://api.pushbullet.com/v2/devices";
    const PUSHES_URL: &str = "https://api.pushbullet.com/v2/pushes";
    const USER_URL: &str = "https://api.pushbullet.com/v2/users/me";

    type Email = String;

    //A subreddit paired with searches received in Json form from the client 
    #[derive(Serialize, Deserialize, PartialEq, Eq, Hash)]
    pub struct SubSearch {
        pub sub: String, 
        pub searches: Vec<String>
    }

    #[derive(Serialize, Deserialize, PartialEq, Eq, Hash)]
    pub struct SubSearches {
        pub subs: Vec<SubSearch>
    }

    #[derive(PartialEq, Eq, Hash)]
    pub struct UserSubSearch{
        pub email: String, 
        pub sub_search: String
    }

    pub struct ReceiverSender {
        pub sender: Mutex<Sender<bool>>,
        pub receiver: Mutex<Receiver<bool>>
    }

    #[allow(dead_code)]
    pub struct SearchThreads {
        pub map: HashMap<Email, ReceiverSender>
    }

    /**
     * Get the devices for a user given their Pushbullet API token
     * @param token - the Pushbullet API token to get devices for 
     * @return - a map of device nicknames to ids 
     */
    pub fn get_devices(token: String) -> HashMap<String, String>{
        let mut devices_map = HashMap::new();
        let client = Client::new();
        //create and send the API request 
        let mut content = client.get(DEVICES_URL)
            .basic_auth::<String, String>(token, None).send().unwrap();
        let content = content.text().unwrap();
        //get the json map from the response
        let json: Value = from_str(&content).unwrap();
        //get the "devices" array in the json 
        let devices = json["devices"].as_array().expect("Could not into array");
        for device in devices{
            let id = device["iden"].as_str().expect("Could not iden");
            let nick = match device["nickname"].as_str(){
                Some(nickname) => nickname,
                None => continue
            };
            devices_map.insert(nick.to_string(), id.to_string());
        }
        devices_map
    }

    /**
     * Send a Pushbullet link to each device provided with the given token
     * @param devices - a Vec containing strings of device ids
     * @param token - the Pushbullet API token to get the devices for 
     * @param url - the url of the Reddit post to send 
     * @param title - the title of the Reddit post to send
     */
    pub fn send_push_link(devices: Vec<String>, token: &str, 
    (url, title): SubResult){
        for device in devices{
            let client = Client::new();
            let mut data = HashMap::new();
            let mut headers = Headers::new();
            data.insert("title", title.to_string());
            data.insert("url", url.to_string());
            data.insert("type", "link".to_string());
            data.insert("device_iden", device.to_string());
            headers.set(ContentType::json());
            headers.set_raw("Access-Token", token);
            client.post(PUSHES_URL).headers(headers).json(&data).send().unwrap();
        }
    }

    /**
     * Gets the name attached to the Pushbullet account 
     * with the provided API token
     * @param token - the API token to get the name for 
     * @return - a string with the name of the Pushbullet user
     */
    pub fn get_user_name(token: &str) -> String {
        let client = Client::new();
        let mut content = client.get(USER_URL)
            .basic_auth::<&str, String>(token, None).send().unwrap();
        let content = content.text().unwrap();
        let json: Value = from_str(&content).unwrap();
        json["name"].as_str().unwrap().to_string()
    }

    /**
     * Gets the email attached to the Pushbullet account 
     * with the provided API token 
     * @param token - the API token to get the email for 
     * @return - a string with the email of the Pushbullet user
     */ 
    pub fn get_email(token: &str) -> String {
        let client = Client::new();
        let mut content = client.get(USER_URL)
            .basic_auth::<&str, String>(token, None).send().unwrap();
        let content = content.text().unwrap();
        let json: Value = from_str(&content).unwrap();
        json["email"].as_str().unwrap().to_string()
    }

    pub fn check_user_results(email: String) {
        let searches = get_searches(email.clone());
        let _interval = get_interval(&email);
        loop{ 
            for search in &searches {
                let sub = &search.sub;
                let query = &search.search;
                let result = get_results(sub.clone(), 
                    query.clone()).unwrap();
                handle_result(&email, result, &search.last_res_url);
            }
            break;
        }
        ()
    }

    pub fn handle_result(_email: &str, (_url, _title): SubResult, _last_result: &str) {

    }
}