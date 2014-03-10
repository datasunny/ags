angular.module('featen.article').factory("Articles", ['$http', 'Alerts', function($http, Alerts) {
        // Get all lists.
        this.getall = function(scall, ecall) {
            var promise = $http.get("/service/articles/");
            var error = {
                type: "warning",
                strong: "Warning!",
                message: "无法获取文章列表，请耍一会再试."
            };
            Alerts.handle(promise, error, undefined, scall, ecall);

            return promise;
        };


        this.getarticle = function(navname, scall, ecall) {
            var promise = $http.get("/service/articles/name/" + navname);
            var error = {
                type: "warning",
                strong: "Warning!",
                message: "无法获取文章列表，请耍一会再试."
            };
            Alerts.handle(promise, error, undefined, scall, ecall);

            return promise;
        };
        
        this.getPageArticles = function(page, scall, ecall) {
        	var promise = $http.get("/service/articles/page/" + page);
        	var error = {
        			type: "warning",
        			strong: "Warning!",
        			message: "Can not fetch articles for current page, please try it later."
        	};
        	Alerts.handle(promise, error, undefined, scall, ecall);
        	return promise;
        };
        
        this.getTotalPageNumber = function(scall, ecall) {
        	var promise = $http.get("/service/articles/totalpage/number");
        	var error = {
        			type: "warning",
        			strong: "Warning!",
        			message: "Can not fetch total page number right now."
        	};
        	Alerts.handle(promise, error, undefined, scall, ecall);
        	return promise;
        };

        this.create = function(data, scall, ecall) {
            var promise = $http.post("/service/articles/", data);
            var error = {
                type: "error",
                strong: "Failed!",
                message: "现在创建不了文章，请等一会再试."
            };
            var success = {
                type: "success",
                strong: "Success!",
                message: "文章创建成功."
            };
            Alerts.handle(promise, error, success, scall, ecall);

            return promise;
        };

        this.save = function(data, scall, ecall) {
            var promise = $http.put("/service/articles/" + data.Id, data);
            var error = {
                type: "info",
                strong: "Failed!",
                message: "暂时无法保存改动."
            };
            Alerts.handle(promise, error, undefined, scall, ecall);

            return promise;
        };

        this.del = function(data, scall, ecall) {
            var promise = $http({
                method: 'DELETE',
                url: "/service/articles/" + data.Id}
            );
            var error = {
                type: "error",
                strong: "Failed!",
                message: "删除文章失败."
            };
            var success = {
                type: "success",
                strong: "Success!",
                message: "成功删除文章."
            };
            Alerts.handle(promise, error, success, scall, ecall);

            return promise;
        };

        return this;
    }]);
