function init(){
	if(typeof Cookies.get('user_id') !== 'undefined'){ 
		$('#loginbutton').text("Logout"); // if user id in cookies, change it to be "Logout"
		$('#loginbutton').attr("href","./"); // dont direct anywhere
		$('#loginbutton').click(function(){ // if clicked
			Cookies.remove("user_id", {path: "/", sameSite: "lax"}); // remove cookie
			var delay = 100; // 0.1 second delay
			setTimeout(function(){ location.reload() }, delay); // refresh
			return false; // dont direct
	});
	}
	else // if no cookie
	{
		$('#viewuser').hide(); // hide viewuser button
	}
}

init() // run init
let data; // init var
$(function() {

    $('#login-form-link').click(function(e) { // if click login form
		$("#login-form").delay(100).fadeIn(100); // fade in
 		$("#register-form").fadeOut(100); // register part fade out
		$('#register-form-link').removeClass('active'); // remove active
		$(this).addClass('active'); // set login as active
		e.preventDefault(); // dont run intended
	});
	$('#register-form-link').click(function(e) { // if click register
		$("#register-form").delay(100).fadeIn(100); // fade in
 		$("#login-form").fadeOut(100); // fade out login
		$('#login-form-link').removeClass('active'); // remove active attrib
		$(this).addClass('active'); // set register as active
		e.preventDefault(); // dont run intended
	});

});

let loginendpoint = endpoints.login + "/login" // set endpoins
let registerendpoint = endpoints.login + "/register"

function getFormData($form){
    // use jquery to serialise form into JSON
    var unindexed_array = $form.serializeArray();
    var indexed_array = {};

    $.map(unindexed_array, function(n, i){
        indexed_array[n['name']] = n['value'];
    });

    return indexed_array;
}


function replaceQR(data){ // inject qr code into html
	let html = `<p>Please scan the following QR Code with Google Authenticator, or TOTP client.</p>
	<img src="data:image/png;base64, ${data.qrdata}" alt="QR Code" /><br>
	<a href="/login.html">Click here to login</a>
	`
	$('#register-form').html(html) .// append
}

$('#register-submit').click(
	function(){ // for register
		$.ajax(registerendpoint, {
        type: "POST",
        //the url where you want to sent the userName and password to
        dataType: 'json',
        async: true,
        xhrFields: {
      withCredentials: true // ajax to register
   },
        //json object to sent to the authentication url
        data: JSON.stringify(getFormData($('#register-form'))), // send json of form
        success: function (data) {
        console.log("Success"); // if okay
        replaceQR(data); // replace qr
        },
        error: function(data){
        	console.log("Register Error") // else error
        	let html = `
        	<p style="color:red;" id="error">${data.responseText}, please try again.</p>
        	`; // send error
        	if ($('#register-form').length == 0){
        		$('#register-form').prepend(html);   // put in html 
        	}
        	else {
        		$('#error').val(data.responseText) // put in html
        	}
        },
    })
	})

$('#login-submit').click(
	function(){ // for login
		data = getFormData($('#login-form')); // get data
		// replace with OTP
		let html = ` 
		<input type="text" name="code" placeholder="OTP Code" value>
		<input type="button" name="verify-submit" id="verify-submit" tabindex="4" class="form-control btn btn-login" value="Submit">
		`
		// append html
		$('#login-form').html(html)
	})
$(document).on('click', '#verify-submit' , function() { // for submit otp
	let data1 = getFormData($('#login-form')); // get email passwd
	data.totp = data1.code //get totp
	$.ajax(loginendpoint, {
       	type: "POST", // ajax to login
        //the url where you want to sent the userName and password to
        async: true,
        xhrFields: {
      withCredentials: true
   },
        //json object to sent to the authentication url
        data: JSON.stringify(data), // send data
        success: function (data) {
        	console.log("Login Success"); // success
        	let html = `
        	<p style="color:green;" id="success">${data}, redirecting to home page.</p>
        	`;
        	if ($('#success').length == 0){ // put html
        		$('#login-form').prepend(html);  
        	}
        	else {
        		$('#success').val(data.responseText) // else error
        	}
        	// Your delay in milliseconds
			var delay = 2000; 
			setTimeout(function(){ window.location = "../"; }, delay); // send to home
        },
        error: function(data){
        	console.log("Login Error") // send error
        	let html = `
        	<p style="color:red;" id="error">${data.responseText}, please try again.</p>
        	`;
        	if ($('#error').length == 0){
        		$('#login-form').prepend(html);   // append html
        	}
        	else {
        		$('#error').val(data.responseText) // append html
        	}
        },
    })
});