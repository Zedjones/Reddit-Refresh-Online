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
use rocket::response::{NamedFile, Redirect};
use rocket::http::{Cookie, Cookies};
use rocket::Data;
use reqwest::Client;
use cookie::SameSite::Lax;
use reqwest::header::{Headers, ContentType};
use std::collections::HashMap;
use std::str;
use serde_json::{Value, from_str};
use reddit_refresh_online::pushbullet::get_user_name;
use reddit_refresh_online::subparser::validate_subreddit;

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

#[derive(Serialize)]
struct JsonValue{
	key: String,
	value: String
}

#[derive(Serialize, Deserialize)]
struct SubSearch {
	sub: String, 
	searches: Vec<String>
}

#[post("/process", format="application/json", data="<sub>")]
fn process(sub: Json<SubSearch>) {
	println!("{}", sub.0.sub);
	println!("{:#?}", sub.0.searches);
}

#[post("/test", data="<var>")]
fn test_route(var: Data) {
	println!("{}", str::from_utf8(var.peek()).unwrap());
}

#[get("/")]
fn index(mut cookies: Cookies) -> Template {
	let mut context = HashMap::new();
	for cookie in cookies.iter() {
		println!("{}", cookie);
	}
	match cookies.get_private("push_token"){
		Some(cookie_token) => {
			let token = cookie_token.to_owned();
			let name = get_user_name(token.value());
			context.insert("login", name)
		}, 
		None => context.insert("login", "Login".to_string()), 
	};
	Template::render("index", context)
}

#[get("/<file..>")]
fn files(file: PathBuf) -> Option<NamedFile> {
    NamedFile::open(Path::new("templates/").join(file)).ok()
}

#[get("/handle_token?<code>")]
fn handle_token(mut cookies: Cookies, code: PushCode) -> Redirect {
	let token = get_token(&code);
	let mut cookie = Cookie::new("push_token", token);
	cookie.set_same_site(Lax);
	cookies.add_private(cookie);
	Redirect::to("/")
}

#[post("/validate_subreddit", data = "<sub>")]
fn validate_route(sub: String) -> Json<JsonValue> {
	let is_valid = validate_subreddit(sub);
	let is_valid = is_valid.unwrap().to_string();
	let result = JsonValue{key: "is_valid".to_string(), value: is_valid};
	Json(result)
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
	rocket::ignite().mount("/", routes![handle_token, index, files, validate_route, test_route, process])
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
