angular.module('featen.deal').factory("Deals", ['$http', 'Alerts', function($http, Alerts) {
        // Get all lists.
        this.getCurrentDeals = function(scall, ecall) {
            var promise = $http.get("/service/deals/");
            var error = {
                type: "warning",
                strong: "Warning!",
                message: "Unable to retrieve all places. Try again in a few minutes."
            };
            Alerts.handle(promise, error, undefined, scall, ecall);

            return promise;
        };


        this.getDeal = function(navname, scall, ecall) {
            var promise = $http.get("/service/deals/" + navname);
            var error = {
                type: "warning",
                strong: "Warning!",
                message: "Unable to retrieve deal information right now."
            };
            Alerts.handle(promise, error, undefined, scall, ecall);

            return promise;
        };
        
        this.getPageDeals = function(pagenumber, scall, ecall) {
            var promise = $http.get("/service/deals/page/" + pagenumber);
            Alerts.handle(promise, undefined, undefined, scall, ecall);
            return promise;
        };

        this.addProductToCart = function(data, scall, ecall) {
            var promise = $http.post("/service/cart/", data);
            var error = {
                type: "warning",
                strong: "Warning!",
                message: "Unable to add product to cart. Try again in a few minutes."
            };
            var success = {
                type: "success",
                strong: "Success!",
                message: "Add product to cart success!"
            };
            Alerts.handle(promise, error, undefined, scall, ecall);

            return promise;
        };
        
 
        
        return this;
}]);

