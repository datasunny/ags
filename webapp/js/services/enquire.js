angular.module('featen.enquire').factory("Enquires", ['$http', 'Alerts', function($http, Alerts) {
    this.searchcount = function(searchtext, scall, ecall) {
      	 var promise = $http.get("/service/enquire/search/" + searchtext + "/count");
           var error = {
               type: "warning",
               strong: "Warning!",
               message: "No response..."
           };
           Alerts.handle(promise, error, undefined, scall, ecall);

           return promise;
      };
      
      this.search = function(searchtext, pagenumber, scall, ecall) {
      	 var promise = $http.get("/service/enquire/search/" + searchtext +"/page/"+pagenumber);
           var error = {
               type: "warning",
               strong: "Warning!",
               message: "No match..."
           };
           Alerts.handle(promise, error, undefined, scall, ecall);

           return promise;
      };    
	
	this.getReviewboardDetail = function( scall, ecall) {
            var promise = $http.get("/service/reviewboard");
            var error = {
                type: "warning",
                strong: "Warning!",
                message: "Unable to get reviewboard detail."
            };
            
            Alerts.handle(promise, error, undefined, scall, ecall);

            return promise;
        };
        
        this.reviewLater = function(data, scall, ecall) {
            var promise = $http.post("/service/reviewboard", data);
            var error = {
                type: "warning",
                strong: "Warning!",
                message: "Unable to add product to reviewboard. Try again in a few minutes."
            };
            
            Alerts.handle(promise, error, undefined, scall, ecall);

            return promise;
        };
        
        this.addEnquire = function (data, scall, ecall) {
            var promise = $http.post("/service/enquire", data);
            var error = {
                type: "warning",
                strong: "Warning!",
                message: "Unable to add enquire right now, Try again in a few minutes."
            };
            
            Alerts.handle(promise, error, undefined, scall, ecall);

            return promise;
        };
        
        this.getEnquiresCountByCond = function(cond, scall, ecall) {
            var promise = $http.get("/service/enquire/count/" + cond);
            var error = {
                type: "warning",
                strong: "Warning!",
                message: "Can not get enquire count right now."
            };
            Alerts.handle(promise, error, undefined, scall, ecall);
            return promise;
        };
        
        
        
        this.getEnquiresByCond = function(cond, scall, ecall) {
            var promise = $http.get("/service/enquire/cond/" + cond);
            var error = {
                type: "warning",
                strong: "Warning!",
                message: "Can not get enquires list right now."
            };
            Alerts.handle(promise, error, undefined, scall, ecall);
            return promise;
        };
        
        
        this.getEnquire = function(data, scall, ecall) {
            var promise = $http.get("/service/enquire/id/"+ data);
            var error = {
                type: "error",
                strong: "Failed!",
                message: "Can not get order now."
            };
            
            Alerts.handle(promise, error, undefined, scall, ecall);
            return promise;
        };
        
        
        this.saveEnquire = function(data, scall, ecall) {
            var promise = $http.put("/service/enquire/id/"+data.Id, data);
            var error = {
                type: "error",
                strong: "Failed!",
                message: "Can not update enquire record，"
            };
            Alerts.handle(promise, error, undefined, scall, ecall);
            return promise;
        };
        return this;
}]);

