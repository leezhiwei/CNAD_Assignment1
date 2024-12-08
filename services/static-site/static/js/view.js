let viewuser = endpoints.login + "/profile"
let updateuser = endpoints.login + "/profile/update"
let delres = endpoints.vehicles + "/cancel"
let modres = endpoints.vehicles + "/modify"
let tierend = endpoints.billing + "/gettier"
let payment = endpoints.billing + "/payment"
let resarr;
let userdata;
let tierarr;

function getFormData($form){
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
    });
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
        userdata = data.user
        console.log("Success");
        console.log(data);
        populate($('#userForm'), data.user)
        tier = tierarr.find(tier => tier.ID === data.user['membership_tier_id']);
        console.log(tier.TierName)
        $('#membership_tier').val(tier.TierName)
        renderReservations(data.rentalHistory)
        resarr = data.rentalHistory;
        $('#canbutton').click(function(){
            let resno = $('#canbutton').attr("resno");
            $.ajax({
                url: delres+`/${resno}`,
                type: "DELETE",
                //the url where you want to sent the userName and password to
                success: function (data) {
                console.log("Success");
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
        $('#updbutton').click(function(){
            let resno = parseInt($('#updbutton').attr("resno"));
            let reser = resarr.find(res => res.reservation_id === resno);
            let start = reser.start_time.slice(0,-1)
            let stop = reser.end_time.slice(0,-1)
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
            $('#updateres').append(html)
            $('#updateresbut').click(function(){
                let formdata = getFormData($('#updateresform'))
                let senddata = {"reservation_id": parseInt(formdata["reservationId"]), "user_id": parseInt(Cookies.get("user_id")), "vehicle_id": parseInt(formdata["vehicleId"]), "start_time": formdata["startTime"].slice(0,-1), "end_time": formdata["endTime"].slice(0,-1)}
                    console.log(senddata)
                    $.ajax(modres, {
                        type: "PUT",
                        //the url where you want to sent the userName and password to
                        async: true,
                        //json object to sent to the authentication url
                        data: JSON.stringify(senddata),
                        success: function (data) {
                        console.log("Success");
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

function gettier(){
    $.ajax('http://localhost:8082/gettier', {
        type: "GET",
        //the url where you want to sent the userName and password to
        dataType: 'json',
        async: true,
        //json object to sent to the authentication url
        success: function (data) {
        console.log("Success");
        tierarr = data;
        },
        error: function(err){
            console.log("Error")
            console.log(err)
        }
    })
}
gettier()
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


$('#checkoutbut').click(function(){
    let formdata = getFormData($('#checkoutForm'))
    // logic here
    if (formdata.reservationid == ''){
        console.log("Payment Error")
            let html = `
            <p style="color:red;" id="error">Reservation ID is blank, please try again.</p>
            `;
            if ($('#error').length == 0){
                $('#checkoutForm').prepend(html);  
            }
            else {
                $('#error').val("Reservation ID is blank, please try again.")
            }
    }
    let reservation = resarr.find(res => res.reservation_id === parseInt(formdata.reservationid));
    if (reservation === undefined){
            console.log("Payment Error")
                let html = `
                <p style="color:red;" id="error">Reservation ID is invalid, please try again.</p>
                `;
                if ($('#error').length == 0){
                    $('#checkoutForm').prepend(html);
                    return  
                }
                else {
                    $('#error').val("Reservation ID is invalid, please try again.")
                    return
                }   
    }
                $('#paymentemail').val(userdata.email)
                let startDateobj = new Date(reservation.start_time)
                let endDateobj = new Date(reservation.end_time)
                let millisec = endDateobj.getTime()-startDateobj.getTime()
                let hours = Math.floor(millisec/(60*60*1000))
                $('#price').val(hours * 40)
                $('#memdiscount').val(hours*40*(tier.DiscRate/100))
                $('#total').val(hours * 40 - (hours*40*(tier.DiscRate/100)))
                $('#checkoutbut').click(function(){
                    let checkoutdata = getFormData($('#checkoutForm'));
                    let paymentdata = $('#pmethod').val();
                    let userid = Cookies.get("user_id");
                    let senddata = {"userid": userid, "resid": checkoutdata.reservationid, "amount": checkoutdata.total ,"pmethod": paymentdata}
                    $.ajax(payment, {
                    type: "POST",
                    //the url where you want to sent the userName and password to
                    async: true,
                    //json object to sent to the authentication url
                    data: JSON.stringify(senddata),
                    success: function(data){
                        sessionStorage.setItem('paymentdetails', JSON.stringify(senddata));
                        window.location = "../success.html"
                    },
                    error: function(err){
                        console.log("Error");
                        console.log(err);
                    }
                });
                });
        })