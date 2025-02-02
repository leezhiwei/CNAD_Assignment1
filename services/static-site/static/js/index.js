let allveh = endpoints.vehicles + '/allvehicles'
let availveh = endpoints.vehicles + '/vehicles' // declare endpoints
let resveh = endpoints.vehicles + '/reserve'
let vehall;
let vehavail; // declare var for storage

function getFormData($form){
    // use jquery to serialise form into JSON
    var unindexed_array = $form.serializeArray();
    var indexed_array = {};

    $.map(unindexed_array, function(n, i){
        indexed_array[n['name']] = n['value'];
    });

    return indexed_array;
}

function vehfleet(data){
    // get in all vehicle fleet data
	let index = 0;
	let html = ``
	let html1 = ``
	vehdata = data
	data.forEach(function(car){
		index += 1
		if (index == 1){ // first selected car, inject HTML
			html1 += `
		<li class="active"><a href="#vehicle-${index}">${car.make} ${car.model}</a><span class="active">&nbsp;</span></li>
		`;
			html += `
	<div class="vehicle-data" id="vehicle-${index}" style="">
            <div class="col-md-6 wow fadeIn animated" data-wow-offset="100" style="visibility: visible;">
                <div class="vehicle-img">
                    <img class="img-responsive" src="data:image/png;base64, ${car.VehiclePicture}" alt="Vehicle">
                </div>
            </div>
            <div class="col-md-3 wow fadeInUp animated" data-wow-offset="200" style="visibility: visible;">
                <div class="vehicle-price">From $40<span class="info"> rent per hour</span></div>
                <table class="table vehicle-features">
                    <tbody><tr>
                        <td>Make</td>
                        <td>${car.make}</td>
                    </tr>
                    <tr>
                        <td>Model</td>
                        <td>${car.model}</td>
                    </tr>
                    <tr>
                        <td>Status</td>
                        <td>${car.status}</td>
                    </tr>
                    <tr>
                        <td>Cleanliness</td>
                        <td>${car.cleanliness}</td>
                    </tr>
                    <tr>
                        <td>Year</td>
                        <td>${car.year}</td>
                    </tr>
                </tbody></table>
                <a href="#teaser" class="reserve-button scroll-to"><span class="glyphicon glyphicon-calendar"></span> Reserve now</a>
            </div>
        </div>
	`
	return; // end fn
		} // subsequent car, inject HTML (not selected)
		html1 += `
		<li><a href="#vehicle-${index}">${car.make} ${car.model}</a><span class="active">&nbsp;</span></li>
		`;
		html += `
	<div class="vehicle-data" id="vehicle-${index}" style="">
            <div class="col-md-6" data-wow-offset="100">
                <div class="vehicle-img">
                    <img class="img-responsive" src="data:image/png;base64, ${car.VehiclePicture}" alt="Vehicle">
                </div>
            </div>
            <div class="col-md-3" data-wow-offset="200" >
                <div class="vehicle-price">From $40<span class="info"> rent per hour</span></div>
                <table class="table vehicle-features">
                    <tbody><tr>
                        <td>Make</td>
                        <td>${car.make}</td>
                    </tr>
                    <tr>
                        <td>Model</td>
                        <td>${car.model}</td>
                    </tr>
                    <tr>
                        <td>Status</td>
                        <td>${car.status}</td>
                    </tr>
                    <tr>
                        <td>Cleanliness</td>
                        <td>${car.cleanliness}</td>
                    </tr>
                    <tr>
                        <td>Year</td>
                        <td>${car.year}</td>
                    </tr>
                </tbody></table>
                <a href="#teaser" class="reserve-button scroll-to"><span class="glyphicon glyphicon-calendar"></span> Reserve now</a>
            </div>
        </div>
	`
	});
	$('#vehnav').append(html1); // use jquery to add to nav bar
	$('#vehicledata').append(html); // also add to vehicle data.
	
}
function availablefleet(data){
	vehavail = data // for all the available fleer
	let html2 = ``
	data.forEach(function (car){ // add it to the dropdown menu
		html2 += `<option value="${car.VehiclePicture}" carid="${car.vehicle_id}">${car.make} ${car.model}</option>`
	})
	$('#car-select').append(html2) // append HTML
}
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
	$.ajax(allveh, { // ajax for all vehicles
        type: "GET",
        //the url where you want to sent the userName and password to
        dataType: 'json',
        async: false,
        //json object to sent to the authentication url
        success: function (data) { // if ajax is okay
        	console.log("Success");
        	vehfleet(data) // run above fn to populate html
        },
        error: function (data) { // if error
        	console.log("Error"); // print error in console
            alert("Error with retrieving data, please refresh the page.") // Alert Window
        	console.log(data) // print out error specifc in console
        },
    });
    $.ajax(availveh, { // ajax for only available vehicle.
        type: "GET",
        //the url where you want to sent the userName and password to
        dataType: 'json',
        async: false,
         //json object to sent to the authentication url
        success: function (data) { // if ajax is okay
        	console.log("Success");
        	availablefleet(data) // run above fn to populate html
        },
        error: function (data) { // if error
        	console.log("Error"); // print error in console
            alert("Error with retrieving data, please refresh the page.") // Alert Window
        	console.log(data) // print out error specifc in console
        },
    })
}

init()  // run init functions

$('#reservebutt').click(function(){ // jquery 
	let formdata = getFormData($('#checkout-form'))
	let senddata = {"user_id": parseInt(Cookies.get('user_id')),"vehicle_id": parseInt(formdata['carid']),"start_time": formdata['pick-up'].slice(0,-1),"end_time": formdata['drop-off'].slice(0,-1)}
	console.log(senddata)
    $.ajax(resveh, {
        type: "POST",
        //the url where you want to sent the userName and password to
        async: true,
        //json object to sent to the authentication url
        data: JSON.stringify(senddata),
        xhrFields: {
        withCredentials: true // Essential for cross-site requests
    },
        success: function (data) { // if success with reserve
        console.log("Success"); // append html
        let html = `
        	<p style="color:green;" id="success">Update Success, redirecting to home page.</p>
        	`;
        	if ($('#success').length == 0){ // if no exist
        		$('#status').prepend(html);  // append
        	}
        	else { // else update.
        		$('#success').val(data.responseText)
        	}
        	// Your delay in milliseconds
			var delay = 2000; 
			setTimeout(function(){ window.location = "../"; }, delay); // send back to home
        },
        error: function(data){ // if error
        	console.log("Update Error") // tell them to try again, append html
        	let html = `
        	<p style="color:red;" id="error">${data.responseText}, please try again.</p>
        	`;
        	if ($('#error').length == 0){
        		$('#status').prepend(html);  
        	}
        	else {
        		$('#userForm').val(data.responseText)
        	}
        },
    })
})