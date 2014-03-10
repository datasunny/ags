angular.module('featen.enquire').controller('ReviewboardController', ['$scope', '$routeParams', '$location', 'Global', 'Enquires', 'User', 'StageData', function($scope, $routeParams, $location, Global, Enquires, User, StageData) {
        $scope.global = Global;
        $scope.data = {};
        
        $scope.getReviewboardDetail = function() {
        	var savedDataId = $routeParams.SavedDataId;
        	if (savedDataId !== undefined && savedDataId !== '' ) {
        		var stagedata = StageData.get(savedDataId);
                if (stagedata !== undefined) {
                    $scope.data = stagedata;
                    StageData.del(savedDataId);
                } else {
                    var r = $location.path().split("/")[1];
                    $location.path("/" + r);
                }
        	} else {
	            Enquires.getReviewboardDetail(function(c) {
	                $scope.data.Reviewboard = c;
	                $scope.data.products = c.Products;
	            });
        	}
        };
        
        $scope.addEnquire = function() {
            if ($scope.global.user == null) {
            	var stageDataId = StageData.add($scope.data);
                var r = $location.path().split("/")[1];
                var redirecturl = "/" + r + "/savedid/" + stageDataId;
            	$location.path("/signin/redirect/"+Base64.encode(redirecturl));
            } else {
            	
                var e = {Products: $scope.data.products,
                    Subject: $scope.data.subject,
                    Message: $scope.data.message};

                Enquires.addEnquire(e, function(o) {
                	$location.path("/");
                });
            }
        };
    }]);
