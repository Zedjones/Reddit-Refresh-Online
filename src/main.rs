#![feature(plugin, custom_derive, decl_macro)]
#![plugin(rocket_codegen)]

extern crate rocket;
extern crate rocket_contrib;
extern crate reqwest;
extern crate serde_json;
extern crate reddit_refresh_online;
extern crate cookie;
extern crate serde;
#[macro_use]
extern crate serde_derive;

use std::path::{Path, PathBuf};
use rocket_contrib::{Template, Json};
use rocket::response::{NamedFile, Redirect, Failure};
use rocket::http::{Cookie, Cookies, Status};
use reqwest::Client;
use cookie::SameSite::Lax;
use reqwest::header::{Headers, ContentType};
use std::collections::HashMap;
use std::str;
use serde_json::{Value, from_str};
use reddit_refresh_online::pushbullet::{get_user_name, get_email};
use reddit_refresh_online::subparser::validate_subreddit;
use reddit_refresh_online::searches_db::searches_db::{add_search, delete_sub_searches};

//Constant declarations for URLs, tokens, etc.
const OAUTH_URL: &str = "https://api.pushbullet.com/oauth2/token";
const CLIENT_ID: &str = "PR0sGjjxNmfu8OwRrawv2oxgZllvsDm1";
const CLIENT_SECRET: &str = "VdoOJb5BVCPNjqD0b02dVrIVZzkVD2oY";
const TOKEN: &str = "o.dlldl3QXAZ1zgfFsAZQyTS673KnNbf2w";

//The Pushbullet code received from the handle_token route 
#[allow(dead_code)]
#[derive(FromForm)]
struct PushCode {
	code: String, 
	state: String
}

//A generic Json value containing a key and value which are both strings 
#[derive(Serialize)]
struct JsonValue{
	key: String,
	value: String
}

//A subreddit paired with searches received in Json form from the client 
#[derive(Serialize, Deserialize)]
struct SubSearch {
	sub: String, 
	searches: Vec<String>
}

/**
 * Route to process a Json body from the client containing a subreddit 
 * as well as an array of search terms 
 * @param sub - a deserialized SubSearch object from the request body
 */
#[post("/process", format="application/json", data="<sub_search>")]
fn process(mut cookies: Cookies, sub_search: Json<SubSearch>) 
-> Result<(), Failure>{
	// TODO: test that this logic actually works
	//grab token from cookies and get the email attached to that token
	let token = cookies.get_private("push_token").unwrap().to_owned();
	let email = get_email(&token.value());
	//delete the previous searches attached to this user and subreddit
	match delete_sub_searches(&email, &sub_search.sub) {
		Ok(_) => (),
		Err(_) => return Err(Failure(Status::NotAcceptable))
	};
	//add all of the new searches 
	for search in &sub_search.searches {
		add_search(&email, &sub_search.sub, &search);
	}
	println!("{}", sub_search.sub);
	println!("{:#?}", sub_search.searches);
	Ok(())
}

/**
 * Get the main page of the website/webserver 
 * @return - A template containing the index Handlebars file 
 * with an applied context
 */
#[get("/")]
fn index(mut cookies: Cookies) -> Template {
	let mut context = HashMap::new();
	//get the private cookie for push_token
	match cookies.get_private("push_token"){
		//ff one exists, set model["login"] to the cookie
		Some(cookie_token) => {
			let token = cookie_token.to_owned();
			let name = get_user_name(token.value());
			context.insert("login", name)
		}, 
		//otherwise, simply supply the default "Login"
		None => context.insert("login", "Login".to_string()), 
	};
	Template::render("index", context)
}

/**
 * A basic file server route to server static content relative to /templates/
 * @param file - a path buffer which contains a path relative to /templates/
 * @return - either a NamedFile with the file or nothing if the file 
 * does not exist
 */
#[get("/<file..>")]
fn files(file: PathBuf) -> Option<NamedFile> {
    NamedFile::open(Path::new("templates/").join(file)).ok()
}

/**
 * Route to handle the Pushbullet code provided by Pushbullet's OAuth
 * @param cookies - cookies object provided by Rocket to add the push_token to
 * @param code - PushCode object containing a state and a OAuth code
 */
#[get("/handle_token?<code>")]
fn handle_token(mut cookies: Cookies, code: PushCode) -> Redirect {
	let token = get_token(&code);
	//create a new cookie called push_token and set it to the token
	let mut cookie = Cookie::new("push_token", token);
	//set same site to lax or else the private cookies will not work
	cookie.set_same_site(Lax);
	cookies.add_private(cookie);
	Redirect::to("/")
}

/**
 * Route to check whether or not a subreddit is valid 
 * @param sub - the subreddit to check for validity
 * @return - a Json value mapping "is_valid" to a boolean
 */
#[post("/validate_subreddit", data = "<sub>")]
fn validate_route(sub: String) -> Json<JsonValue> {
	let is_valid = validate_subreddit(sub);
	let is_valid = is_valid.unwrap().to_string();
	let result = JsonValue{key: "is_valid".to_string(), value: is_valid};
	Json(result)
}

/**
 * Function to get a Pushbullet token given an OAuth code 
 * @param code - the PushCode containing the code 
 * @param string - the Pushbullet token 
 */
fn get_token(code: &PushCode) -> String {
	let client = Client::new();
	let mut content = client.post(OAUTH_URL);

	//create data map for the request 
	let mut data = HashMap::new();
	data.insert("client_secret", CLIENT_SECRET);
	data.insert("code", &code.code);
	data.insert("grant_type", "authorization_code");
	data.insert("client_id", CLIENT_ID);

	//create headers to specify content type and client access token
	let mut headers = Headers::new();
	headers.set(ContentType::json());
	headers.set_raw("Access-Token", TOKEN);

	content.headers(headers).json(&data);

	let content = content.send().unwrap().text().unwrap();
	let json: Value = from_str(&content).unwrap();

	json["access_token"].as_str().unwrap().to_string()
}

fn main() {
	rocket::ignite().mount("/", routes![handle_token, index, files, validate_route, process])
		.attach(Template::fairing()).launch();
}

#[cfg(test)]
mod tests{

	extern crate rocket;
	extern crate rocket_contrib;
	extern crate reddit_refresh_online;
	use {validate_route, JsonValue};
	use rocket_contrib::Json;

	#[test]
	fn test_validate_route() {
		//test values and correct return values
		let tests = ["Battlefield", "mechanicalkeyboards", "doggesawe", "the_donald"];
		let correct_res = ["true", "true", "false", "true"];
		//iterate using map and collect results from calling route
		let actual_res: Vec<Json<JsonValue>> 
			= tests.iter().map(|sub| validate_route(sub.to_string())).collect();
		//grab the inner json value returned from the route
		let actual_res: Vec<String>
		 	= actual_res.into_iter().map(|json_val| json_val.into_inner().value).collect();
		//assert that the actual results are the correct results 
		assert_eq!(actual_res, correct_res);
	}
}
