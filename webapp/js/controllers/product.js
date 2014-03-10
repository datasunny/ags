angular.module('featen.product').controller('ProductsController', ['$scope', '$routeParams', '$location', 'Global', 'Products', function ($scope, $routeParams, $location, Global, Products) {
    
    $scope.searchtext = "";
    $scope.searchcount = {};
    $scope.currPage = 1;
    $scope.totalPageNumber = 1;
    
    $scope.search = function() {
    	$scope.currPage = 1;
    	var t = $scope.searchtext;
    	if ($scope.searchtext.length == 0) 
    		t = " ";
    	Products.searchcount(t, function(sc) {
    		$scope.searchcount = sc;
    		$scope.totalPageNumber = Math.ceil(sc.Total / sc.PageLimit);
    	});
    	Products.search(t, 1, function(cs) {
    		$scope.products = cs;
    	});
    };
    
    $scope.setpage = function(n) {
    	var t = $scope.searchtext;
    	if ($scope.searchtext.length == 0) 
    		t = " ";
    	Products.search(t, n, function(cs) {
        	$scope.currPage = n;
    		$scope.products = cs;
    	});
    };
    
    $scope.getStatusName = function(s) {
        switch (s) {
        case 0:
            return "Not For Sale";
        case 1:
            return "For Sale";
        case 2:
            return "On Sale";
        }
    };
}]);

angular.module('featen.product').controller('ProductEditController', ['$scope', '$routeParams', '$location', 'Global','StageData','Products', function ($scope, $routeParams, $location, Global,StageData, Products) {
    $scope.data = {};
    var navname = $routeParams.NavName;
    var savedDataId = $routeParams.SavedDataId;
    var uploadedUrlsId = $routeParams.UploadedUrls;

    $scope.getProduct = function() {
        if (savedDataId !== undefined && savedDataId !== '') {
            var stagedata = StageData.get(savedDataId);
            if (stagedata !== undefined) {
                $scope.data = stagedata;
                StageData.del(savedDataId);
            } else {
                Products.getproduct(navname, function(p) {
                    $scope.data = p;
                });
            }
            if (uploadedUrlsId !== undefined && uploadedUrlsId !== '') {
                //$scope.data.Photos = StageData.get(uploadedUrlsId).split(";");
            	$scope.data.Photos = $scope.data.Photos.concat(StageData.get(uploadedUrlsId).split(";"));
                $scope.data.CoverPhoto = $scope.data.Photos[0];
                StageData.Del(uploadedUrlsId);
            }
        } else {
                Products.getproduct(navname, function(p) {
                    $scope.data = p;
                    if (p.SaleURL !== undefined && p.SaleURL !== null) {
                        for (x in p.SaleURL) {
                            var url = p.SaleURL[x];
                            if (url.search("taobao.com")>0)
                                $scope.data.taobaourl = url;
                            if (url.search("weibo.com")>0)
                                $scope.data.weibourl = url;
                            if (url.search("wechat.com")>0)
                                $scope.data.wechaturl = url;
                            if (url.search("ebay.com")>0)
                                $scope.data.ebayurl = url;
                        }
                    }
                });
        } 
    };


    $scope.saveproduct = function() {
        var price = parseFloat($scope.data.Price);
        var discount = parseFloat($scope.data.Discount);
        $scope.data.Price = price;
        $scope.data.Discount = discount;

        var pstatus = parseInt($scope.data.Status);
        $scope.data.Status = pstatus;

        $scope.data.SaleURL = [];
        if ($scope.data.taobaourl !== undefined && $scope.data.taobaourl !== '') {
            $scope.data.SaleURL.push($scope.data.taobaourl);
        }
        if ($scope.data.wechaturl !== undefined && $scope.data.wechaturl !== '') {
            $scope.data.SaleURL.push($scope.data.wechaturl);
        }
        if ($scope.data.ebayurl !== undefined && $scope.data.ebayurl !== '') {
            $scope.data.SaleURL.push($scope.data.ebayurl);
        }
        Products.saveproduct($scope.data, function(c) {
            $location.path("/products");
        });
    };

    $scope.jumptoupload = function() {
        var stageDataId = StageData.add($scope.data);
        var r = $location.path().split("/");
        var redirecturl = "/" + r[1] + "/"+ r[2] + "/savedid/" + stageDataId;
        $location.path('/uploadfile/redirect/'+Base64.encode(redirecturl));
    };

    
}]);
angular.module('featen.product').controller('ProductAddController', ['$scope', '$routeParams', '$location', 'Global','StageData','Products', function ($scope, $routeParams, $location, Global,StageData, Products) {
    //$scope.data = {"CoverPhoto":"/images/product_logo.jpg"};
    $scope.data = {};
	
    var savedDataId = $routeParams.SavedDataId;
    var uploadedUrlsId = $routeParams.UploadedUrls;
    $scope.getNewProduct = function() {
        if (savedDataId !== undefined && savedDataId !== '') {
            var stagedata = StageData.get(savedDataId);
            if (stagedata !== undefined) {
                $scope.data = stagedata;
                StageData.del(savedDataId);
            } else {
                //$scope.data = {"CoverPhoto":"/images/product_logo.jpg"};
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


    $scope.addproduct = function() {
        var price = parseFloat($scope.data.Price);
        var discount = parseFloat($scope.data.Discount);
        $scope.data.Price = price;
        $scope.data.Discount = discount;

        var pstatus = parseInt($scope.data.Status);
        $scope.data.Status = pstatus;

        $scope.data.SaleURL = [];
        if ($scope.data.taobaourl !== undefined && $scope.data.taobaourl !== '') {
            $scope.data.SaleURL.push($scope.data.taobaourl);
        }
        if ($scope.data.wechaturl !== undefined && $scope.data.wechaturl !== '') {
            $scope.data.SaleURL.push($scope.data.wechaturl);
        }
        if ($scope.data.ebayurl !== undefined && $scope.data.ebayurl !== '') {
            $scope.data.SaleURL.push($scope.data.ebayurl);
        }
        Products.add($scope.data, function(c) {
            $location.path("/products");
        });
    };

    $scope.jumptoupload = function() {
        var stageDataId = StageData.add($scope.data);
        var r = $location.path().split("/")[1];
        var redirecturl = "/" + r + "/savedid/" + stageDataId;
        $location.path('/uploadfile/redirect/'+Base64.encode(redirecturl));
    };

    
}]);
