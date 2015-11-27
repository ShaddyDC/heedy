
//login attempts to log into ConnectorDB. If successful, it refreshes the site. if not, it notifies the user.
function login() {
	usrname = $("#username").val()
	pass = $("#password").val()

	if (usrname == "") {
		alert("Please type in a username!");
	} else if (pass=="") {
		alert("Please type in a password!");
	} else {
		//While waiting for response, disable the inputs
		$("#username").prop('disabled', true);
		$("#password").prop('disabled', true);

		//Set the auth header
		authHeader = "Basic " + btoa(usrname + ":" + pass);

		$.ajax({
			type: "GET",
			xhrFields: {
				withCredentials: true
			},
			url: "/api/v1/login?remember=true",
			success: function(data) {
					location.reload(true);
			},
			error: function(request, textStatus, errorThrown) {
				$("#msform").addClass("has-error")
				$("#username").prop('disabled', false);
				$("#password").prop('disabled', false);
				$("#password").val("");
				$("#password").focus();
			},
			beforeSend: function (xhr) {
		        xhr.setRequestHeader('Authorization', 'Basic ' + btoa(usrname + ":" + pass));
		    }
		});


	}

	return false;
}
