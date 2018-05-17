//raw HTML as strings, needs to be converted into template format for use with
// HTML/template later. To make code injection safe.

package main


const PageTop =`
<!DOCTYPE html>
<html lang="en">
<head>
  <title>ndevi</title>
   <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css">
  <script src="https://ajax.googleapis.com/ajax/libs/jquery/1.11.3/jquery.min.js"></script>
  <script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/js/bootstrap.min.js"></script>
</head>
<body>
	
<div class="container-fluid">
<a href="http://www.ndevi.com/" style="text-decoration:none;"><h2>ndevi</h2></a>
	  		
	<div class="col-xs-12 col-md-12">
`




const PageBot = `
    
   
	</div>
	 <br></br>	
	<footer id="footer">	
    <p>ndevi solutions 2015. Contact <a href ="mailto:support@ndevi.com">support@ndevi.com</a> </p>    
    </footer>	
	
</div>
</body>
</html>	
`

const PageTopCent = `
<!DOCTYPE html>
<html lang="en">
<head>
  <title>ndevi</title>
   <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css">
  <script src="https://ajax.googleapis.com/ajax/libs/jquery/1.11.3/jquery.min.js"></script>
  <script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/js/bootstrap.min.js"></script>
</head>
<body>
	
<div class="container-fluid">
	  		
	<div class="col-xs-12 col-md-12">
		<div class="col-xs-12 col-md-4">
		</div>
			<div class="col-xs-12 col-md-4">
				<br></br><br></br><br></br>
				<a href="http://www.ndevi.com/" style="text-decoration:none;"><h2>ndevi</h2></a>
					

`


const PageBotCent =`
			<form action="/signin" method="post">
	<div class="form-group">
    <label for="exampleInputEmail1">Email address</label>
    <input type="email" name ="exampleInputEmail1" class="form-control" id="exampleInputEmail1" placeholder="first.last@abcinsurance.com" required>
	  </div>
  	<div class="form-group">
    <label for="exampleInputPassword1">ndevi password</label>
    <input type="password" name = "exampleInputPassword1"  class="form-control" id="exampleInputPassword1" placeholder="Password" required>
  	</div>
  	
  	<button type="submit" class="btn btn-success">Login</button>
  		
	<a href="/reslink">Reset or change password</a>
	
	</form>	
	<br>

			<p>Please do not enter your password if you don't see the lock sign
  	in the address bar or if the address starts with anything <b>except</b> https://www.ndevi.com  (secure http only).</p>

			
			<br></br>	
				<footer id="footer">	
    			<p>ndevi solutions 2015. Contact <a href ="mailto:support@ndevi.com">support@ndevi.com</a> </p>    
    			</footer>	
	

			</div>
		<div class="col-xs-12 col-md-4">
		</div>
	</div>
	</div>
</body>
</html>	


`

const index = `				
		<div class="col-xs-12 col-md-9">
	<h4>Technology assisted negotiation for reinsurance</h4>
	<p>ndevi is a B2B startup that provides alternatives to conventional reinsurance placement 
	negotiations 
		with technology-assisted bargaining solutions. 
	 These placement process resembles "auction-like" mechanisms that mimic or improve upon the outcomes from
	 conventional negotiations. These mechanisms are intended to follow "sealed bids" formats to 
	 preserve privacy of proprietary business information. We also adopt randomization as a tool
	 to influence bidder incentives. The insurance industry has been gradually working towards 
	 automating underwriting processes. As these solutions become robust, they can be integrated with
	 automated negotiations and lead to more efficient matching of needs and capacities.</p> 
	 				

	
	
	<h4>Why do this?</h4>
	
	<p>Reinsurance is a highly cyclical industry. 
	If there is an excess supply of
	capacity (the current case in most markets), then prices will be low regardless of the procurement mechanism. 
	</p>
	<ul>
	<li>
	Our objective is to create systems that can, at the very least, reliably replicate the outcomes of conventional negotiations.
	
		</li>
	
		
	<li>
	Partially automated negotiations will reduce time and costs associated with formal negotiations. When accumulated at the firm level, 
	these savings can be significant for insurers, reinsurers and brokers alike. 

	</li>
	
	<li>
	Gradually moving from PDF and other formats to web-based forms and data interchange will make it easier for insurers and reinsurers to better
	integrate their negotiation processes with their evolving business intelligence and analytics processes.
	</li>
	
	
	<li>
	By selectively revealing information from bids, auction like mechanisms can help manage the impact of information asymmetries.
	</li>

		
	<li>Business lines requiring joint reinsurance from multiple reinsurers (aviation) and placements in the retro market can particularly benefit from
	variants of multi-unit auctions, which are perfectly suited for the purpose (just like auctions of Treasury securities).  
	</li>
	
	<li>
	Typically, reinsurers do not learn much from negotiations 
	as to why they did, or did not, win a business. Auctions generate information. 
	Although proprietary bidding data cannot be revealed, disclosure of some aggregate market level
	auction data will be beneficial to all parties, including regulators. 
	</li>

	
	<li> We view low cost, partially automated negotiations as an essential component in creating virtual 
	reinsurance exchanges. In India, the industry regulator is trying to create such an exchange, 
	though its incentives are more inclined towards keeping track of contracts and settlements. In fact, 
	the regulator has stated that participation will be made mandatory. The proposal is silent on
	negotiation frameworks.</li>
	
	<li>We have no intentions of bypassing reinsurance brokers. Some
	insurers and reinsurers prefer to deal through brokers, while some don't. Brokers provide valuable services, 
	in both procurement as well as claims management. They also provide expert advice and data analytics services.
	We think automated negotiations will be useful for brokers as well, though the may choose different auction 
	rules, as their earnings are tied to prices.
	</li>
				
	</ul>

<h4>What we are not doing</h4>
	<p>Sometimes people get the incorrect impression that we are building an open electronic reinsurance exchange
	where buyers and sellers post their demands or capacities. We 
	have absolutely no intention of making an insurer's reinsurance requests public. Our platform should
	be viewed as a collection of private virtual negotiation rooms. Each insurer will decide which
	reinsurers to invite to participate in its search process. These details will not be public.</p>
	
	
<h4>Timing</h4>	
	<p>Currently, there is an excess of capacity in global reinsurance markets and retrocession markets.  
	Consequently, in most markets, reinsurance prices are quite competitive at present. 
	A period of low prices may seem like an odd time to build negotiation platforms. However, this is actually a
	great time. For a variety of technical, theoretical, and organizational difficulties, it is hard to design appropriate matching algorithms for reinsurance. A period of
	low prices implies the opportunity costs of learning and experimentation are also low. This gives us time
	to refine our methods, make them robust, and earn the industry's support and confidence. There are indications 
	that demand is recovering in some markets already. As this recovery strengthens, we'll be in a solid
	position to provide value. </p> 
	
	
	
	<h4>People</h4>
	<p><b>Full time:</b> Shreemoy Mishra (founder) is an engineer turned economist based in New Delhi, India. 
	He completed his doctoral work in economics at the University of Texas at Austin (2008) where he also received a MS in operations 
	research & industrial engineering. 	
	Post PhD, he served on the economics faculty at Oberlin College in Ohio, Vassar College in New York, and 
	 most recently, at IIIT Delhi, a computer science and communications engineering research university. 
	 
	His research explores applications 
	of game theory in credit and insurance markets.
	He left academics in July 2015 to focus full time on ndevi. 
	</p>	
	
	
	<p><b>Part time:</b> Bani Dhir is group leader, benefits practice, at Mercer Consulting
	in New Delhi. Bani was previously at Munich Re in Germany and Genpact in India. She is currently wrapping up
	some obligations associated with her current position, and plans to focus on starting up full time in January. 
	</p>
	
	
	<h4>Current Status</h4>

	
	
	<p> 	
	Thus far, we have had discussions with half a dozen insurance companies.  
	Given the relationship heavy nature of the business, 
	some of the people we talk to are naturally skeptical. However, many people we talk to at 
	insurance companies are cooperative and curious. 


	We are currently developing prototypes of these platforms targeting three types of negotiation
	environments. These will help   
	identify sources of apprehension, and discover new problems. 
	
	</p> 
	
	
	</div>


	<div class="col-xs-12 col-md-3">

	<h4>Sign in</h4>
	
	<form action="/signin" method="post">
	<div class="form-group">
    <label for="exampleInputEmail1">Email address</label>
    <input type="email" name ="exampleInputEmail1" class="form-control" id="exampleInputEmail1" placeholder="first.last@abcinsurance.com" required>
	  </div>
  	<div class="form-group">
    <label for="exampleInputPassword1">ndevi password</label>
    <input type="password" name = "exampleInputPassword1"  class="form-control" id="exampleInputPassword1" placeholder="Password" required>
  	</div>
  	
  	<button type="submit" class="btn btn-success">Login</button>
  	
  	<a href="/reslink">Reset or change password</a>

	</form>	
		
	<p>We use temporary session cookies for user authentication. Cookies are automatically deleted upon log out or when
	users quit or restart a browser.  
	If you see the message 
	"<i>securecookie: the value is not valid</i>, then please restart your browser to clear the
	cookies. 
	 </p>
	


	<h4>New user registration </h4>
	<form action="/regreq" method="post">
	<div class="form-group">
    <label for="exampleInputEmail1">Work Email address</label>
    <input type="email" name  = "exampleInputEmail1" class="form-control" id="exampleInputEmail1" placeholder="first.last@abcinsurance.com" required>
	</div>
  	<button type="submit" class="btn btn-primary">Register</button>		
	</form>	
	
	</div>
`

const retry = `
<div class="col-xs-12 col-md-3">
<form action="/signin" method="post">
	<div class="form-group">
    <label for="exampleInputEmail1">Email address</label>
    <input type="email" name ="exampleInputEmail1" class="form-control" id="exampleInputEmail1" placeholder="first.last@abcinsurance.com" required>
	  </div>
  	<div class="form-group">
    <label for="exampleInputPassword1">ndevi password</label>
    <input type="password" name = "exampleInputPassword1"  class="form-control" id="exampleInputPassword1" placeholder="Password" required>
  	</div>
  	
  	<button type="submit" class="btn btn-success">Login</button>
  		
	<a href="/reslink">Reset or change password</a>
	
	</form>	
	<br>

</div>
`

const register = `
	<h4>New User Registration </h4>
		
	<p>If we detect a prior registration with your email ID, then we do not send any emails. 
	Please use the password reset option.</p>
	
	<p>Please copy and paste the single use "token" sent to your email. Then choose a new password 
	consisting of at least 8 characters. You may use any combination of letters, numbers, and symbols. 
	Passwords are case sensitive.
	 </p>
	
	<p>If you did not receive the email, please check your spam folder (mark it "not spam"!).</p>

	
	<form action="/signup" method="post">
	<div class="form-group">
    <label for="exampleInputEmail1">Work Email address</label>
    <input type="email" name  = "exampleInputEmail1" class="form-control" id="exampleInputEmail1" placeholder="first.last@abcinsurance.com" required>
	  </div>
	
	<div class="form-group">
			    <label> Firm or Employer type </label> 
  				<div class="input-group" >
  				<label class="radio-inline"> 
  				<input type="radio" name="RadioProd" id="inlineRadio1" value="Ins" required> Insurer
				</label>
				<label class="radio-inline"> 
  				<input type="radio" name="RadioProd" id="inlineRadio2" value="RIns" required> Reinsurer
				</label>
				<label class="radio-inline">
  				<input type="radio" name="RadioProd" id="inlineRadio3" value="Broker" required> Broker
				</label>				
  				</div>
  		</div>		
	
	<div class="form-group">
    <label for="token">Enter token received by email</label>    
    <input type="password" name = "token" class="form-control" id="token" placeholder="Token" required>
  	</div>
	
	<p> Please use strong passwords. If you don't feel like selecting a password at this time, 
	just copy and paste a section of the token we sent you.  However, if you use
	this strategy, then please <b>do not write</b> it down anywhere! You can
	use the password reset option to get a new token every time, though this can get tiresome.</p>

  	<div class="form-group">
    <label for="exampleInputPassword1">Choose a password</label>
    <input type="password" name = "exampleInputPassword1" class="form-control" id="exampleInputPassword1" placeholder="Password" minlength="8" required>
  	</div>
  	
  	<div class="form-group">
    <label for="exampleInputPassword2">Enter password again</label>
    
    <input type="password" name = "exampleInputPassword2" class="form-control" id="exampleInputPassword2" placeholder="Password" minlength="8" required>
  	</div>
  	
  	    	<button type="submit" class="btn btn-primary">Register</button>	
		
		
	</form>	

`

const emailReg = `
<h4>New user registration </h4>
	<form action="/regreq" method="post">
	<div class="form-group">
    <label for="exampleInputEmail1">Work Email address</label>
    <input type="email" name  = "exampleInputEmail1" class="form-control" id="exampleInputEmail1" placeholder="first.last@abcinsurance.com" required>
	</div>
  	<button type="submit" class="btn btn-primary">Register</button>	
		
		
	</form>	
`

const emailPwdResetForm = `
<div class="col-xs-12 col-md-3">
<h4>Reset or change password </h4>
	<form action="/resetreq" method="post">
	<div class="form-group">
    <label for="exampleInputEmail1">Work Email address</label>
    <input type="email" name  = "exampleInputEmail1" class="form-control" id="exampleInputEmail1" placeholder="first.last@abcinsurance.com" required>
	</div>
  	<button type="submit" class="btn btn-warning">Reset</button>		
	</form>	
	<br></br>
</div>

`





const emailReset = `
<h4>Reset Password </h4>
	<p>Please enter (or copy and paste) the single use "token" sent to your email. 
	If you did not receive the email, please check your spam folder (mark it "not spam"!). 
	 </p>
	
	<p>Choose a new password 
	consisting of at least 8 characters. You may use any combination of letters, numbers, and symbols. 
	Passwords are case sensitive.</p>
 
	 
	<form action="/resetpwd" method="post">
	<div class="form-group">
    <label for="exampleInputEmail1">Work Email address</label>
    <input type="email" name  = "exampleInputEmail1" class="form-control" id="exampleInputEmail1" placeholder="first.last@abcinsurance.com" required>
	  </div>
	
	<div class="form-group">
    <label for="token">Enter token received by email</label>    
    <input type="password" name = "token" class="form-control" id="token" placeholder="Token" required>
  	</div>
	
	<p> We encrypt all passwords before storing them in the database. Please use strong passwords with <b>8 or more</b> characters, preferably a combination of
	upper/lower case letters, numbers, and symbols. 
	You can always reset forgotten passwords via email.</p> 
	<p>If you don't feel like selecting a 
	strong password at this time, just copy and paste a section of the token we sent you.  If you use
	this strategy, then please <b>do not write the password</b> down anywhere. You can
	use the password reset option to get a new token every time. This will make your
	account less susceptible to email phishing attacks.</p>
  	<div class="form-group">
    <label for="exampleInputPassword1">Choose a password</label>
    <input type="password" name = "exampleInputPassword1" class="form-control" id="exampleInputPassword1" placeholder="Password" minlength="8" required>
  	</div>
  	
  	<div class="form-group">
    <label for="exampleInputPassword2">Enter password again</label>
    
    <input type="password" name = "exampleInputPassword2" class="form-control" id="exampleInputPassword2" placeholder="Password" minlength="8" required>
  	</div>
  	  		
    <button type="submit" class="btn btn-primary">Reset</button>	
		
		
	</form>	
`


//The following are templates for creating RFQ's
const newRFQ = `
<div class="col-xs-12 col-md-8">
<h4>Create new request for quotes</h4>

<p>If you have any documents to attach, you may do so on the following page. </p>

 <form action="/request" method = "post">
 	<div class="form-group">
    <label for="client">Client Name</label>    
    <input type="text" name = "client" class="form-control" id="client" placeholder="Client Name" required>
  	</div>
  	
  	<div class="form-group">
    <label for="cat">Category</label>    
    <input type="text" name = "cat" class="form-control" id="cat" placeholder="Group Life, Group Health, Fire, Liability "required >
  	</div>
	
	<div class="form-group">
			    <label> Current or new customer </label>
  				<div class="input-group" >
  				<label class="radio-inline"> 
  				<input type="radio" name="RadioProd" id="inlineRadio1" value="Renewal" required> Renewal
				</label>
				<label class="radio-inline"> 
  				<input type="radio" name="RadioProd" id="inlineRadio2" value="New Quote" required> New client
				</label>
  				</div>
  	</div>	
 	
 	<div class="form-group">
    <label for="RIType">Type of reinsurance</label>    
    <input type="text" name = "RItype" class="form-control" id="RItype" placeholder="Excess of loss, proportional" required>
  	</div>
  	<div class="form-group">
    <label for="IV">Sum Insured (INR)</label>    
    <input type="text" name = "sumIns" class="form-control" id="sumIns" placeholder="1,50,00,000" required>
  	</div>
  	
  	<div class="form-group">
    <label for="reten">Retention (INR)</label>    
    <input type="text" name = "reten" class="form-control" id="reten" placeholder="50,00,000" required>
  	</div> 	
	
	<p>Enter reinsurers' emails separated by commas. You can enter this list later. Please
	be careful, we cannot verify the correctness of email Ids you enter.</p>
	<div class="form-group">
    <label for="RIcs">Reinsurers Emails</label>    
    <input type="text" name = "RIemails" class="form-control" id="list" placeholder="xyz@abcRE.com, abc@xyzRe.com," >
  	</div> 	
	
	<div class="form-group">
    <label for="remarks">Additional notes for recipients</label> 	
 	<textarea class="form-control" name = "remarks" rows="10" placeholder = "Limit 1450 characters, about 250 words" 
 		maxlength = "1450"></textarea>
 	</div>

		
	
 <button type="submit" class="btn btn-primary">Create</button>
	

 </form>
 </div>

 <br></br>
`
