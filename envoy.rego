package envoy

default allow = false

allow {
	is_post
	is_dogs
	claims.username == "alice"
}

is_post {
	input.attributes.request.http.method == "POST"
}

is_dogs {
    input.attributes.request.http.path == "/pets/dogs"
}

claims := payload {
	io.jwt.verify_hs256(bearer_token, "B41BD5F462719C6D6118E673A2389")

	[_, payload, _] := io.jwt.decode(bearer_token)
}

bearer_token := t {
	v := input.attributes.request.http.headers.authorization
	startswith(v, "Bearer ")
	t := substring(v, count("Bearer "), -1)
}
