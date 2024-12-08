function init(){
	let jsonstring = sessionStorage.getItem("paymentdetails"); // get paymentdetail
	if (jsonstring === null) { // if no storage
		$('html').html(`<h1 style="color: red;">Unable to access payment data, please try again.</h1>`)
		window.location = '../' // inject and send back
	}
	let paymentobj = JSON.parse(jsonstring); // get json in
	$('#text').append(`<p>You have paid $${paymentobj.amount}, via ${paymentobj.pmethod}</p>`);
	// add to html
	$('#text').append(`<p>Please check View User tab to check Membership Points and Membership Status.</p><br>`);
	// then send back home
	var delay = 5000; 
	setTimeout(function(){ window.location = "../"; }, delay);
}

init()