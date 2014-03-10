angular.module('featen.share').factory("Shares", ['$http', 'Alerts', function($http, Alerts) {
        this.uploadPhoto = function(data, scall, ecall) {
                var xhr = new XMLHttpRequest();
                xhr.addEventListener("load", scall, false);
                xhr.addEventListener("error", ecall, false);
                xhr.open("POST", "/service/uploadphoto");
                var r = xhr.send(data);
                return r;
            
            var error = {
                type: "error",
                strong: "Failed!",
                message: "Cannot upload right now"
            };
            var success = {
                type: "success",
                strong: "Success!",
                message: "Upload success..."
            };
        };

        return this;
    }]);



