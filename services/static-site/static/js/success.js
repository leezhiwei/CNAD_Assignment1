function init(){
	let jsonstring = sessionStorage.getItem("paymentdetails");
	if (jsonstring === null) {
		$('html').html(`<h1 style="color: red;">Unable to access payment data, please try again.</h1>`)
		window.location = '../'
	}
	let paymentobj = JSON.parse(jsonstring);
	$('#text').append(`<p>You have paid $${paymentobj.amount}, via ${paymentobj.pmethod}</p>`);
	$('#text').append(`<p>Please check View User tab to check Membership Points and Membership Status.</p><br>`);
	var delay = 5000; 
	setTimeout(function(){ window.location = "../"; }, delay);
}

init()