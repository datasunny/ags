angular.module('featen.system').factory("Global", [function() {
        var _this = this;
        _this._data = {
            user: window.user,
            employee: window.employee,
            authenticated: !!window.user,
            //employee: false
            isemployee: !!window.employee,
            notemployee: !window.employee
        };

        return _this._data;
    }]);

angular.module('featen.system').factory("StageData", [function() {

        // for ProductAddController
        var StageData = [];
        var currIndex = 0;
        this.get = function(id) {
            var data;
            angular.forEach(StageData, function(d) {
                if (d.id === parseInt(id))
                    data = d.data;
            });
            return data;
        };
        this.add = function(adddata) {
            var i = currIndex++;
            StageData.push({id: i, data: adddata});
            return i;
        };
        this.del = function(id) {
            var oldStageData = StageData;
            StageData = [];
            angular.forEach(oldStageData, function(d) {
                if (d.id !== parseInt(id))
                    StageData.push(d);
            });
        };
        
        return this;
    }]);