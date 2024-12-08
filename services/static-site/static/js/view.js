let viewuser = endpoints.login + "/profile"
let updateuser = endpoints.login + "/profile/update"
let delres = endpoints.vehicles + "/cancel"
let modres = endpoints.vehicles + "/modify"
let tierend = endpoints.billing + "/gettier" // endpoints
let payment = endpoints.billing + "/payment"
let resarr;
let userdata;
let tierarr;

function getFormData($form){
    // use jquery to serialise form into JSON
    var unindexed_array = $form.serializeArray();
    var indexed_array = {};

    $.map(unindexed_array, function(n, i){
        indexed_array[n['name']] = n['value'];
    });

    return indexed_array;
}


function renderReservations(reservations) {
    const reservationsBody = $('#reservationsBody');
    reservationsBody.empty(); // Clear any existing rows

    reservations.forEach(reservation => {
        // for each reservation. append html
        const row = `
            <tr>
                <td>${reservation.reservation_id}</td>
                <td>${reservation.vehicle_id}</td>
                <td>${new Date(reservation.start_time).toLocaleString()}</td>
                <td>${new Date(reservation.end_time).toLocaleString()}</td>
                <td>${reservation.status}</td>
                <td><button type="button" id="canbutton" resno="${reservation.reservation_id}">Cancel</button></td>
                <td><button type="button" id="updbutton" resno="${reservation.reservation_id}">Update</button></td>
            </tr>
        `;
        reservationsBody.append(row);
    // append back.
    });
}

function populate(frm, data) {
    // append json to frm
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
        $('#loginbutton').text("Logout"); // if user id in cookies, change it to be "Logout"
        $('#loginbutton').attr("href","./"); // dont direct anywhere
        $('#loginbutton').click(function(){ // if clicked
            Cookies.remove("user_id", {path: "/", sameSite: "lax"}); // remove cookie
            var delay = 100; // 0.1 second delay
            setTimeout(function(){ location.reload() }, delay); // refresh
            return false; // dont direct
    });
    }
	else{
        // set entire html
		let html = `
		<h1 style="color:red;">User not logged in, please log in.<h1>
		`
		$('html').html(html)
		var delay = 1000; 
		setTimeout(function(){ window.location = "../"; }, delay); // redir home
	}
	$.ajax(viewuser, { // get user data
        type: "GET",
        //the url where you want to sent the userName and password to
        dataType: 'json',
        async: true,
    xhrFields: {
      withCredentials: true
   },
        //json object to sent to the authentication url
        success: function (data) {
        userdata = data.user // store in global
        console.log("Success");
        populate($('#userForm'), data.user) // populate form
        tier = tierarr.find(tier => tier.ID === data.user['membership_tier_id']); // find specific tier
        $('#membership_tier').val(tier.TierName) // set in form
        renderReservations(data.rentalHistory) // put into reservarion table
        resarr = data.rentalHistory; // set global
        $('#canbutton').click(function(){ // cancel
            let resno = $('#canbutton').attr("resno"); // get the number
            $.ajax({
                url: delres+`/${resno}`,
                type: "DELETE", // ajax delete
                //the url where you want to sent the userName and password to
                success: function (data) {
                console.log("Success");
                // inject html okay
                let html = `
                    <p style="color:green;" id="success">Successfully delete reservation, redirecting to home page.</p>
                    `;
                    if ($('#success').length == 0){
                        $('#userForm').prepend(html);  
                    }
                    else {
                        $('#success').val(data.responseText)
                    }
                    // Your delay in milliseconds
                    var delay = 2000; 
                    setTimeout(function(){ window.location = "../"; }, delay); // return home
                },
                error: function(data){
                    console.log("Update Error")
                    // error, put error
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
        $('#updbutton').click(function(){ //update
            let resno = parseInt($('#updbutton').attr("resno")); // get resid
            let reser = resarr.find(res => res.reservation_id === resno); // get object
            let start = reser.start_time.slice(0,-1) // remove Z from timestring
            let stop = reser.end_time.slice(0,-1) 
            // inject html
            let html = `
            <form id="updateresform">
            <div class="form-group">
                <label for="reservationId">Reservation ID</label>
                <input type="text" name="reservationId" id="reservationId" value="${reser.reservation_id}" readonly>
            </div>
            <div class="form-group">
                <label for="vehicleId">Vehicle ID</label>
                <input type="text" name="vehicleId" id="vehicleId" value="${reser.vehicle_id}" readonly>
            </div>
            <div class="form-group">
                <label for="startTime">Start Time</label>
                <input type="datetime-local" name="startTime" id="startTime" value="${start}">
            </div>
            <div class="form-group">
                <label for="endTime">End Time</label>
                <input type="datetime-local" name="endTime" id="endTime" value="${stop}">
            </div>
            <div class="form-group">
                <label for="status" >Status</label>
                <input type="text" name="status" id="status" value="${reser.status}" readonly>
            </div>
            <button type="button" id="updateresbut">Update</button>
        </form>
        `
            $('#updateres').append(html) // append to html
            $('#updateresbut').click(function(){ // onclick
                let formdata = getFormData($('#updateresform')) // getdata
                // put into format
                let senddata = {"reservation_id": parseInt(formdata["reservationId"]), "user_id": parseInt(Cookies.get("user_id")), "vehicle_id": parseInt(formdata["vehicleId"]), "start_time": formdata["startTime"].slice(0,-1), "end_time": formdata["endTime"].slice(0,-1)}
                    // ajax
                    $.ajax(modres, {
                        type: "PUT",
                        //the url where you want to sent the userName and password to
                        async: true,
                        //json object to sent to the authentication url
                        data: JSON.stringify(senddata), // send data
                        success: function (data) {
                        console.log("Success");
                        // if success inject html, green
                        let html = `
                            <p style="color:green;" id="success">Update Success, redirecting to home page.</p>
                            `;
                            if ($('#success').length == 0){
                                $('#updateresform').prepend(html);  
                            }
                            else {
                                $('#success').val(data.responseText)
                            }
                            // Your delay in milliseconds
                            var delay = 2000; 
                            setTimeout(function(){ window.location = "../"; }, delay);
                        },
                        // else inject error html
                        error: function(data){
                            console.log("Update Error")
                            let html = `
                            <p style="color:red;" id="error">${data.responseText}, please try again.</p>
                            `;
                            if ($('#error').length == 0){
                                $('#updateresform').prepend(html);  
                            }
                            else {
                                $('#updateresform').val(data.responseText)
                            }
                        },
                    })
                            })
                        });
                        },
    })
}

function gettier(){ // ajax get tiers
    $.ajax('http://localhost:8082/gettier', {
        type: "GET",
        //the url where you want to sent the userName and password to
        dataType: 'json',
        async: true,
        //json object to sent to the authentication url
        success: function (data) {
        console.log("Success");
        tierarr = data; // store globally
        },
        error: function(err){
            console.log("Error")
            console.log(err)
        }
    })
}
gettier() // get tiers
init() // then init



$('#updatebutton').click(function(){ // update user
	let formdata = getFormData($('#userForm')) // get details
	let senddata = {"email": formdata["email"], "phone": formdata["phone"]} // put in format
	$.ajax(updateuser, { // ajax
        type: "POST", 
        //the url where you want to sent the userName and password to
        async: true,
        //json object to sent to the authentication url
        data: JSON.stringify(senddata), // send data
        xhrFields: {
        withCredentials: true // Essential for cross-site requests
    },
        success: function (data) {
        console.log("Success");
        // if success inject success html
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
			setTimeout(function(){ window.location = "../"; }, delay); // redir home
        },
        // else error inject error html
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


function genInvoice(userdata, checkoutdata, tierarr){
    // generate invoice
    const { jsPDF } = window.jspdf;
    var doc = new jsPDF(); // new pdf object
    doc.setFontSize(20); // set size
    tier = tierarr.find(tier => tier.ID === userdata.membership_tier_id); // get specific tier
    // inject data
    let text = `
    Email: ${userdata.email}
    User ID: ${userdata.user_id}
    Phone Number: ${userdata.phone}
    Tier: ${tier.TierName}
    Points: ${userdata.membership_point}
    Amount Paid: $${checkoutdata.amount}
    Paid Using: ${checkoutdata.pmethod}`

    doc.text(20, 10 + (1 * 10), text); // add text
    doc.save('Invoice.pdf') // download
}

$('#checkoutbut').click(function(){ // checkout button
    let formdata = getFormData($('#checkoutForm'))  // get data
    // logic here
    if (formdata.reservationid == ''){ // if no reservationid blank
        console.log("Payment Error")
            let html = `
            <p style="color:red;" id="error">Reservation ID is blank, please try again.</p>
            `; // error
            if ($('#error').length == 0){
                $('#checkoutForm').prepend(html);  
            }
            else { // set html
                $('#error').val("Reservation ID is blank, please try again.")
            }
    }
    // try get reservation
    let reservation = resarr.find(res => res.reservation_id === parseInt(formdata.reservationid));
    if (reservation === undefined){ // no reservation found
            console.log("Payment Error")
            // html error
                let html = `
                <p style="color:red;" id="error">Reservation ID is invalid, please try again.</p>
                `;
                if ($('#error').length == 0){
                    $('#checkoutForm').prepend(html);
                    return  
                }
                // update html
                else {
                    $('#error').val("Reservation ID is invalid, please try again.")
                    return
                }   
    }
                // else if found
                $('#paymentemail').val(userdata.email) // fill in email
                let startDateobj = new Date(reservation.start_time)
                let endDateobj = new Date(reservation.end_time) // calculate duration
                let millisec = endDateobj.getTime()-startDateobj.getTime()
                let hours = Math.floor(millisec/(60*60*1000))
                $('#price').val(hours * 40) // add price
                $('#memdiscount').val(hours*40*(tier.DiscRate/100)) // discount
                $('#total').val(hours * 40 - (hours*40*(tier.DiscRate/100))) // payable
                $('#checkoutbut').click(function(){ // if clicked again
                    let checkoutdata = getFormData($('#checkoutForm'));
                    let paymentdata = $('#pmethod').val(); // get data
                    let userid = Cookies.get("user_id");
                    // put format
                    let senddata = {"userid": userid, "resid": checkoutdata.reservationid, "amount": checkoutdata.total ,"pmethod": paymentdata}
                    // make payment to server
                    $.ajax(payment, {
                    type: "POST",
                    //the url where you want to sent the userName and password to
                    async: true,
                    //json object to sent to the authentication url
                    data: JSON.stringify(senddata),
                    success: function(data){ // if ok
                        sessionStorage.setItem('paymentdetails', JSON.stringify(senddata));// set data for success
                        genInvoice(userdata, senddata, tierarr)// make new pdf 
                        window.location = "../success.html" // send to success
                    },
                    error: function(err){ // if errpr
                        console.log("Error"); // print error
                        console.log(err);
                    }
                });
                });
        })