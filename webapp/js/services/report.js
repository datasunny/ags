angular.module('featen.report').factory("Reports", ['$http', 'Alerts', function($http, Alerts) {
        // Get all lists.
        this.getdata = function(cond, scall, ecall) {
            var promise = $http.get("/service/report/" + cond);
            var error = {
                type: "warning",
                strong: "Warning!",
                message: "现在无法查询统计数据."
            };
            Alerts.handle(promise, error, undefined, scall, ecall);

            return promise;
        };

        
        return this;
}]);



