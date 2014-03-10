angular.module('featen.customer').factory("Customers", ['$http', 'Alerts', function($http, Alerts) {
        // Get all lists.
        this.getall = function(cond, scall, ecall) {
            var promise = $http.get("/service/customers/" + cond);
            var error = {
                type: "warning",
                strong: "Warning!",
                message: "无法查询客户信息，请稍后再试."
            };
            Alerts.handle(promise, error, undefined, scall, ecall);

            return promise;
        };
        
        this.searchcount = function(searchtext, scall, ecall) {
        	 var promise = $http.get("/service/customers/search/" + searchtext + "/count");
             var error = {
                 type: "warning",
                 strong: "Warning!",
                 message: "No response..."
             };
             Alerts.handle(promise, error, undefined, scall, ecall);

             return promise;
        };
        
        this.search = function(searchtext, pagenumber, scall, ecall) {
        	 var promise = $http.get("/service/customers/search/" + searchtext +"/page/"+pagenumber);
             var error = {
                 type: "warning",
                 strong: "Warning!",
                 message: "No match..."
             };
             Alerts.handle(promise, error, undefined, scall, ecall);

             return promise;
        };

        this.add = function(data, scall, ecall) {
            var promise = $http.post("/service/customers/", data);
            var error = {
                type: "error",
                strong: "Failed!",
                message: "无法创建新客户，请稍后再试."
            };
            var success = {
                type: "success",
                strong: "Success!",
                message: "客户创建成功."
            };
            Alerts.handle(promise, error, success, scall, ecall);
            return promise;
        };
        
        this.getcustomer = function(id, scall, ecall) {
          var promise = $http.get("/service/customers/id/"+id);
          var error = {
                type: "warning",
                strong: "Warning!",
                message: "无法获取该客户信息，请稍后再试."
            };
            Alerts.handle(promise, error, undefined, scall, ecall);

            return promise;
        };
        
        this.savecustomer = function(data, scall, ecall) {
            var promise = $http.post("/service/customers/id/", data);
            var error = {
                type: "error",
                strong: "Failed!",
                message: "无法保存修改客户资料，请稍后再试."
            };
            var success = {
                type: "success",
                strong: "Success!",
                message: "客户资料保存成功."
            };
            Alerts.handle(promise, error, success, scall, ecall);
            return promise;
        };
        
        this.addcustomerlog = function(data, scall, ecall) {
            var promise = $http.post("/service/customers/log/", data);
            var error = {
                type: "error",
                strong: "Failed!",
                message: "无法添加客户记录，请稍后再试."
            };
            var success = {
                type: "success",
                strong: "Success!",
                message: "客户记录添加成功."
            };
            Alerts.handle(promise, error, success, scall, ecall);
            return promise;
        };
        return this;
}]);

