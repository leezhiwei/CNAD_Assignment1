let viewuser = endpoints.login + "/profile"
let updateuser = endpoints.login + "/profile/update"
let tiers = {
	1: "Basic",
	2: "Premium",
	3: "VIP"
}
function getFormData($form){
    var unindexed_array = $form.serializeArray();
    var indexed_array = {};

    $.map(unindexed_array, function(n, i){
        indexed_array[n['name']] = n['value'];
    });

    return indexed_array;
}


function populate(frm, data) {   
    $.each(data, function(key, value) {  
        var ctrl = $('[name='+key+']', frm);  
        switch(ctrl.prop("type")) { 
            case "radio": case "checkbox":   
                ctrl.each(function() {
                    if($(this).attr('value') == value) $(this).attr("checked",value);
                });   
                break;  
            default:
                ctrl.val(value); 
        }  
    });  
}

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
	else{
		let html = `
		<h1 style="color:red;">User not logged in, please log in.<h1>
		`
		$('html').html(html)
		var delay = 1000; 
		setTimeout(function(){ window.location = "../"; }, delay);
	}
	$.ajax(viewuser, {
        type: "GET",
        //the url where you want to sent the userName and password to
        dataType: 'json',
        async: true,
    xhrFields: {
      withCredentials: true
   },
        //json object to sent to the authentication url
        success: function (data) {
        console.log("Success");
        console.log(data);
        populate($('#userForm'), data.user)
        console.log(tiers[data.user['membership_tier_id']])
        $('#membership_tier').val(tiers[data.user['membership_tier_id']])
        },
    })
}

init()

$('#updatebutton').click(function(){
	let formdata = getFormData($('#userForm'))
	let senddata = {"email": formdata["email"], "phone": formdata["phone"]}
	$.ajax(updateuser, {
        type: "POST",
        //the url where you want to sent the userName and password to
        async: true,
        //json object to sent to the authentication url
        data: JSON.stringify(senddata),
        xhrFields: {
        withCredentials: true // Essential for cross-site requests
    },
        success: function (data) {
        console.log("Success");
        let html = `
        	<p style="color:green;" id="success">Update Success, redirecting to home page.</p>
        	`;
        	if ($('#success').length == 0){
        		$('#userForm').prepend(html);  
        	}
        	else {
        		$('#success').val(data.responseText)
        	}
        	// Your delay in milliseconds
			var delay = 2000; 
			setTimeout(function(){ window.location = "../"; }, delay);
        },
        error: function(data){
        	console.log("Update Error")
        	let html = `
        	<p style="color:red;" id="error">${data.responseText}, please try again.</p>
        	`;
        	if ($('#error').length == 0){
        		$('#userForm').prepend(html);  
        	}
        	else {
        		$('#userForm').val(data.responseText)
        	}
        },
    })
})