function init(){
	if(typeof Cookies.get('user_id') !== 'undefined'){
		$('#loginbutton').text("Logout");
		$('#loginbutton').attr("href","./");
		$('#loginbutton').click(function(){
			Cookies.remove("user_id", {path: "/", sameSite: "lax"});
			var delay = 100; 
			setTimeout(function(){ location.reload() }, delay);
			return false;
	});
	}
	else
	{
		$('#viewuser').hide();
	}
}

init()
let data;
$(function() {

    $('#login-form-link').click(function(e) {
		$("#login-form").delay(100).fadeIn(100);
 		$("#register-form").fadeOut(100);
		$('#register-form-link').removeClass('active');
		$(this).addClass('active');
		e.preventDefault();
	});
	$('#register-form-link').click(function(e) {
		$("#register-form").delay(100).fadeIn(100);
 		$("#login-form").fadeOut(100);
		$('#login-form-link').removeClass('active');
		$(this).addClass('active');
		e.preventDefault();
	});

});

let loginendpoint = endpoints.login + "/login"
let registerendpoint = endpoints.login + "/register"

function getFormData($form){
    var unindexed_array = $form.serializeArray();
    var indexed_array = {};

    $.map(unindexed_array, function(n, i){
        indexed_array[n['name']] = n['value'];
    });

    return indexed_array;
}

function replaceQR(data){
	let html = `<p>Please scan the following QR Code with Google Authenticator, or TOTP client.</p>
	<img src="data:image/png;base64, ${data.qrdata}" alt="QR Code" /><br>
	<a href="/login.html">Click here to login</a>
	`
	$('#register-form').html(html)
}

$('#register-submit').click(
	function(){
		$.ajax(registerendpoint, {
        type: "POST",
        //the url where you want to sent the userName and password to
        dataType: 'json',
        async: true,
        xhrFields: {
      withCredentials: true
   },
        //json object to sent to the authentication url
        data: JSON.stringify(getFormData($('#register-form'))),
        success: function (data) {
        console.log("Success");
        replaceQR(data);
        },
        error: function(data){
        	console.log("Register Error")
        	let html = `
        	<p style="color:red;" id="error">${data.responseText}, please try again.</p>
        	`;
        	if ($('#register-form').length == 0){
        		$('#register-form').prepend(html);  
        	}
        	else {
        		$('#error').val(data.responseText)
        	}
        },
    })
	})

$('#login-submit').click(
	function(){
		data = getFormData($('#login-form'));
		let html = `
		<input type="text" name="code" placeholder="OTP Code" value>
		<input type="button" name="verify-submit" id="verify-submit" tabindex="4" class="form-control btn btn-login" value="Submit">
		`
		$('#login-form').html(html)
	})
$(document).on('click', '#verify-submit' , function() {
	let data1 = getFormData($('#login-form'));
	data.totp = data1.code
	$.ajax(loginendpoint, {
       	type: "POST",
        //the url where you want to sent the userName and password to
        async: true,
        xhrFields: {
      withCredentials: true
   },
        //json object to sent to the authentication url
        data: JSON.stringify(data),
        success: function (data) {
        	console.log("Login Success");
        	let html = `
        	<p style="color:green;" id="success">${data}, redirecting to home page.</p>
        	`;
        	if ($('#success').length == 0){
        		$('#login-form').prepend(html);  
        	}
        	else {
        		$('#success').val(data.responseText)
        	}
        	// Your delay in milliseconds
			var delay = 2000; 
			setTimeout(function(){ window.location = "../"; }, delay);
        },
        error: function(data){
        	console.log("Login Error")
        	let html = `
        	<p style="color:red;" id="error">${data.responseText}, please try again.</p>
        	`;
        	if ($('#error').length == 0){
        		$('#login-form').prepend(html);  
        	}
        	else {
        		$('#error').val(data.responseText)
        	}
        },
    })
});