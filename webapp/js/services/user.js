angular.module('featen.user').factory("User", ['$http', 'Alerts', function($http, Alerts) {
        this.data = {
			user: null,
			authenticated: false,
                        employee: false
                    };
	
	// Get the currently logged in user.
	this.getall = function(call) {
			var promise = $http.get("/service/users/all");
			var error = {
					type: "warning",
					strong: "Warning!",
					message: "Unable to retrieve user information."
			};
			Alerts.handle(promise, error, undefined, call);
			
	};
    this.get = function(call) {
        var promise = $http.get("/service/users/");
        
        Alerts.handle(promise, undefined, undefined, call);
    };
    
    
    this.updateinfo = function(data, scall, ecall) {
        var promise = $http.put("/service/users/", data);
        var error = {
            type: "error",
            strong: "Failed!",
            message: "Could not update. Try again in a few minutes."
        };
        var success = {
            type: "success",
            strong: "Success!",
            message: "Update success."
        };
        Alerts.handle(promise, error, success, scall, ecall);

        return promise;
    };
    
    this.updatepassword = function(data, scall, ecall) {
        var promise = $http.put("/service/users/password", data);
        var error = {
            type: "error",
            strong: "Failed!",
            message: "Could not change password now."
        };
        Alerts.handle(promise, error, undefined, scall, ecall);
        return promise;
    };
    
    this.signin = function(data, scall, ecall) {
        var promise = $http.post("/service/users/signin", data);
        var error = {
            type: "error",
            strong: "Failed!",
            message: "Could not sign in. Try again in a few minutes."
        };
        var success = {
            type: "success",
            strong: "Success!",
            message: "Sign in success."
        };
        Alerts.handle(promise, error, undefined, scall, ecall);

        return promise;
    };
    this.signup = function(data, scall, ecall) {
    	var promise = $http.post("/service/users", data);
    	//var promise = $http({ method: "PUT", url: "/users:8080" });
        var error = {
                type: "error",
                strong: "Failed!",
                message: "Could not sign up. Try again in a few minutes."
            };
            var success = {
                type: "success",
                strong: "Success!",
                message: "Sign up success, redirect to Sign In page."
            };
            Alerts.handle(promise, error, success, scall, ecall);

        return promise;
    };
    this.signout = function(data, scall, ecall) {
    	var promise = $http.post("/service/users/signout", data);
    	//var promise = $http({ method: "PUT", url: "/users:8080" });
        var error = {
                type: "error",
                strong: "Failed!",
                message: "Could not signout. Try again in a few minutes."
            };
            var success = {
                type: "success",
                strong: "Success!",
                message: "Sign out success."
            };
            Alerts.handle(promise, error, undefined, scall, ecall);

        return promise;
    };
    this.sendRecoverMail = function(data, scall, ecall) {
    	var promise = $http.post("/service/recover", data);
        var error = {
                type: "error",
                strong: "Failed!",
                message: "找不到您输入的Email，请检查."
            };
            var success = {
                type: "success",
                strong: "Success!",
                //message: "Recover mail sent to your email address, please follow the instructions."
                message: "恢复密码邮件已经发送到你的邮箱，请查收。"
            };
            Alerts.handle(promise, error, success, scall, ecall);

        return promise;
    };
    this.getShippingOptions = function(scall, ecall) {
            var promise = $http.get("/service/users/shippings");
            var error = {
                type: "warning",
                strong: "Warning!",
                message: "Unable to purchase right now, Try again in a few minutes."
            };
            
            Alerts.handle(promise, error, undefined, scall, ecall);

            return promise;
        };
        
    this.updateAddress = function(data, scall, ecall) {
        var promise = $http.post("/service/users/address", data);
        var error = {
                type: "error",
                strong: "Failed!",
                message: "Update registration address failed，please check."
            };
            var success = {
                type: "success",
                strong: "Success!",
                //message: "Recover mail sent to your email address, please follow the instructions."
                message: "Update registration address success."
            };
            Alerts.handle(promise, error, success, scall, ecall);

        return promise;
    };
    this.getUserOrders = function(scall, ecall) {
        var promise = $http.get("/service/order/user");
        var error = {
                type: "error",
                strong: "Failed!",
                message: "Get user orders failed."
            };
            
            Alerts.handle(promise, error, undefined, scall, ecall);

        return promise;
    };
    
    this.getUserTrans = function(scall, ecall) {
        var promise = $http.get("/service/tran/user");
        var error = {
                type: "error",
                strong: "Failed!",
                message: "Get user trans failed."
            };
            
            Alerts.handle(promise, error, undefined, scall, ecall);

        return promise;
    };
    
    return this;
}]);

