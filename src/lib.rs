#[macro_use]
extern crate diesel;
extern crate dotenv;
extern crate reqwest;
extern crate serde_json;

pub mod schema;
pub mod models;
mod searches_db;

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

    pub fn validate_subreddit(mut sub: String) -> Result<bool, String> {

        //in case subreddit is missing r/
        if !sub.contains("r/"){
            sub = format!("r/{}", sub);
        }

        sub = sub.trim().to_string();

        let url = format!("https://www.reddit.com/{}/\
        about.json", sub);

        let client = Client::new();

        let content = match client.get(&url).send().unwrap().text(){
            Ok(content) => {
                content
            }, 
            Err(e) => {
                let error = format!("Error retrieving webpage: {}", e.to_string());
                return Err(error);
            }
        };

        let json: Value = from_str(&content).unwrap();

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
    use super::subparser::SubResult;

    const DEVICES_URL: &str = "https://api.pushbullet.com/v2/devices";
    const PUSHES_URL: &str = "https://api.pushbullet.com/v2/pushes";
    const USER_URL: &str = "https://api.pushbullet.com/v2/users/me";

    pub fn get_devices(token: String) -> HashMap<String, String>{
        let mut devices_map = HashMap::new();
        let client = Client::new();
        let mut content = client.get(DEVICES_URL)
            .basic_auth::<String, String>(token, None).send().unwrap();
        let content = content.text().unwrap();
        let json: Value = from_str(&content).unwrap();
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

    pub fn get_user_name(token: &str) -> String {
        let client = Client::new();
        let mut content = client.get(USER_URL)
            .basic_auth::<&str, String>(token, None).send().unwrap();
        let content = content.text().unwrap();
        let json: Value = from_str(&content).unwrap();
        json["name"].as_str().unwrap().to_string()
    }
}