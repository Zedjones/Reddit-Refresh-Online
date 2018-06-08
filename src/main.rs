#![feature(plugin)]
#![plugin(rocket_codegen)]
#![feature(custom_derive)]

extern crate rocket;

#[derive(FromForm)]
struct PushCode {
	code: String
}

#[get("/handle_token?<code>")]
fn handle_token(code: PushCode) -> String {
	code.code
}

fn main() {
	rocket::ignite().mount("/", routes![handle_token]).launch();
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
