angular.module('featen.system').factory("Alerts", ['$rootScope', function($rootScope) {
        this.alerts = [];
        var self = this;

        // add places a new alert at the top of the list of alerts. A
        // timeout is made that removes the alert after 15 seconds has
        // expired.
        this.add = function(type, strong, message) {
            this.alerts.unshift({
                type: type,
                strong: strong,
                message: message
            });
            window.setTimeout(function() {
                $rootScope.$apply(self.remove(self.alerts.length - 1));
            }, 5000);
        };

        // remove gets rid of the specified alert.
        this.remove = function(index) {
            this.alerts.splice(index, 1);
        };

        // handle is a helper function for REST calls. The given promise
        // will have a success and error function added to it. Should the
        // call succeed, the success alert will be added and the scall
        // function will be called with the data, statue, headers, and
        // config. Should the call fail, the error alert will be
        // added and the ecall function will be called with the data,
        // statue, headers, and config.
        this.handle = function(promise, error, success, scall, ecall) {
            promise
                    .success(function(data, status, headers, config) {
                        if (success !== undefined) {
                            self.add(success.type, success.strong,
                                    success.message);
                        }

                        if (scall !== undefined) {
                            scall(data, status, headers, config);
                        }
                    })
                    .error(function(data, status, headers, config) {
                        if (error !== undefined) {
                            self.add(error.type, error.strong,
                                    error.message);
                        }

                        if (ecall !== undefined) {
                            ecall(data, status, headers, config);
                        }
                    });
        };
        return this;
    }]);

