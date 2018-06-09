#![feature(plugin)]
#![plugin(rocket_codegen)]
#![feature(custom_derive)]

extern crate rocket;
extern crate rocket_contrib;
extern crate reqwest;
extern crate serde_json;
extern crate reddit_refresh_online;

use std::path::{Path, PathBuf};
use rocket_contrib::Template;
use rocket::response::NamedFile;
use reqwest::Client;
use reqwest::header::{Headers, ContentType};
use std::collections::HashMap;
use serde_json::{Value, from_str};
use reddit_refresh_online::pushbullet::get_user_name;

const OAUTH_URL: &str = "https://api.pushbullet.com/oauth2/token";
const CLIENT_ID: &str = "PR0sGjjxNmfu8OwRrawv2oxgZllvsDm1";
const CLIENT_SECRET: &str = "VdoOJb5BVCPNjqD0b02dVrIVZzkVD2oY";
const TOKEN: &str = "o.dlldl3QXAZ1zgfFsAZQyTS673KnNbf2w";

#[allow(dead_code)]
#[derive(FromForm)]
struct PushCode {
	code: String, 
	state: String
}

#[get("/")]
fn index() -> Template {
	let mut context = HashMap::new();
	context.insert("login", "Test");
	Template::render("index", context)
}

#[get("/<file..>")]
fn files(file: PathBuf) -> Option<NamedFile> {
    NamedFile::open(Path::new("templates/").join(file)).ok()
}

#[get("/handle_token?<code>")]
fn handle_token(mut code: PushCode) -> String {
	code.code = code.code.replace("&state=", "");
	let token = get_token(&code);
	let name = get_user_name(&token);
	name
}

fn get_token(code: &PushCode) -> String {
	let client = Client::new();
	let mut content = client.post(OAUTH_URL);

	let mut data = HashMap::new();
	data.insert("client_secret", CLIENT_SECRET);
	data.insert("code", &code.code);
	data.insert("grant_type", "authorization_code");
	data.insert("client_id", CLIENT_ID);

	let mut headers = Headers::new();
	headers.set(ContentType::json());
	headers.set_raw("Access-Token", TOKEN);

	content.headers(headers).json(&data);

	let content = content.send().unwrap().text().unwrap();
	let json: Value = from_str(&content).unwrap();

	json["access_token"].as_str().unwrap().to_string()
}

fn main() {
	rocket::ignite().mount("/", routes![handle_token, index, files])
		.attach(Template::fairing()).launch();
}

#[cfg(test)]
mod tests{

	extern crate rocket;
	use rocket::local::Client;

	#[test]
	fn test_handle_token(){
		let rocket = rocket::ignite().mount("/", routes![super::handle_token]);
		let client = Client::new(rocket).unwrap();
		let mut result = client.get("/handle_token?code=amfEasdksak").dispatch();
		assert_eq!(result.body_string(), Some("amfEasdksak".into()));
	}
}
