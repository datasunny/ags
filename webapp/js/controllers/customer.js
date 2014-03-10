angular.module("featen.customer").controller("CustomerAddController", ["$scope", "$routeParams", "$location", "Global", "StageData", "Customers", function($scope, $routeParams, $location, Global, StageData,Customers) {
	$scope.data = {};
	
	var savedDataId = $routeParams.SavedDataId;
    var uploadedUrlsId = $routeParams.UploadedUrls;
    $scope.getNewCustomer = function() {
        if (savedDataId !== undefined && savedDataId !== '') {
            var stagedata = StageData.get(savedDataId);
            if (stagedata !== undefined) {
                $scope.data = stagedata;
                StageData.del(savedDataId);
            } else {
                var r = $location.path().split("/")[1];
                $location.path("/" + r);
            }
            if (uploadedUrlsId !== undefined && uploadedUrlsId !== '') {
                $scope.data.Photos = StageData.get(uploadedUrlsId).split(";");
                $scope.data.CoverPhoto = $scope.data.Photos[0];
                StageData.del(uploadedUrlsId);
            }
        } 
    };
    
    $scope.jumptoupload = function() {
        var stageDataId = StageData.add($scope.data);
        var r = $location.path().split("/")[1];
        var redirecturl = "/" + r + "/savedid/" + stageDataId;
        $location.path('/uploadfile/redirect/'+Base64.encode(redirecturl));
    };
    
    $scope.addcustomer = function() {
        Customers.add($scope.data, function(c) {
            $location.path("/");
        });
    };

}]);


angular.module("featen.customer").controller("CustomersController", ["$scope", "$routeParams", "$location", "Global", "Customers", function($scope, $routeParams, $location, Global, Customers) {
        $scope.searchtext = "";
        $scope.searchcount = {};
        $scope.currPage = 1;
        $scope.totalPageNumber = 1;
        
        $scope.search = function() {
        	$scope.currPage = 1;
        	var t = $scope.searchtext;
        	if ($scope.searchtext.length == 0) 
        		t = "@";
        	Customers.searchcount(t, function(sc) {
        		$scope.searchcount = sc;
        		$scope.totalPageNumber = Math.ceil(sc.Total / sc.PageLimit);
        	});
        	Customers.search(t, 1, function(cs) {
        		$scope.customers = cs;
        	});
        };
        
        $scope.setpage = function(n) {
        	var t = $scope.searchtext;
        	if ($scope.searchtext.length == 0) 
        		t = "@";
        	Customers.search(t, n, function(cs) {
            	$scope.currPage = n;
        		$scope.customers = cs;
        	});
        };
    }]);

angular.module("featen.customer").controller("CustomerEditController", ["$scope", "$routeParams", "$location", "Global", "StageData", "Customers", "Shares", function($scope, $routeParams, $location, Global, StageData, Customers, Shares) {
        $scope.global = Global;
        $scope.data = {};
        var id = $routeParams.Id;

        var savedDataId = $routeParams.SavedDataId;
        var uploadedUrlsId = $routeParams.UploadedUrlsId;

        $scope.getcustomer = function() {
            if (savedDataId !== undefined && savedDataId !== '') {
                var stagedata = StageData.get(savedDataId);
                if (stagedata !== undefined) {
                    $scope.customer = stagedata;
                    StageData.del(savedDataId);
                } else {
                    Customers.getcustomer(id, function(d) {
                        $scope.customer = d;
                    });
                }
                if (uploadedUrlsId !== undefined && uploadedUrlsId !== '') {
                    $scope.customer.CoverPhoto = StageData.get(uploadedUrlsId);
                    StageData.del(uploadedUrlsId);
                }
            } else {
                Customers.getcustomer(id, function(d) {
                    $scope.customer = d;
                });
            }
        };

        $scope.jumptoupload = function() {
            var stageDataId = StageData.add($scope.customer);
            var r = $location.path().split("/");
            var redirecturl = "/" + r[1] + "/" + r[2] +"/savedid/" + stageDataId;
            $location.path('/uploadfile/redirect/' + Base64.encode(redirecturl));
        };

        $scope.savecustomer = function() {
            var c = $scope.customer;
            Customers.savecustomer({
                Id: id,
                Name: c.Name,
                CoverPhoto: c.CoverPhoto,
                Phone: c.Phone,
                Email: c.Email,
                Desc: c.Desc
            }, function() {
                $scope.getcustomer();
            });
        };
        $scope.addlog = function() {
            Customers.addcustomerlog({
                CustomerId: id,
                OperationType: $scope.data.OperationType,
                OperationDetail: $scope.data.OperationDetail
            }, function() {
                $scope.getcustomer();
                $scope.data = {};
            });
        };
    }]);
