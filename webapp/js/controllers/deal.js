angular.module('featen.deal').controller('DealsController', ['$scope', '$routeParams', '$location', 'Global', 'Deals', 'Enquires', function ($scope, $routeParams, $location, Global, Deals, Enquires) {
    $scope.global = Global;
    $scope.data = {taobaourl:"",ebayurl:""};
    $scope.findDeals = function() {
        Deals.getCurrentDeals(function(ps) {
            $scope.deals = ps;
        });

    };
    /*----------------------------------------------------*/
    /*	Flexslider
    /*----------------------------------------------------*/
    var loadSlider = function() {   
    $('#intro-slider').flexslider({
          namespace: "flex-",
          controlsContainer: "",
          animation: 'fade',
          controlNav: false,
          directionNav: true,
          smoothHeight: true,
          slideshowSpeed: 7000,
          animationSpeed: 600,
          randomize: false,
       });
    };
       
    $scope.load = function() {
        var navname = $routeParams.NavName;
        Deals.getDeal(navname, function(p) {
            $scope.deal = p;
            $scope.photos = p.Photos;
            setTimeout( loadSlider, 1);
            
            if (p.SaleURL !== undefined && p.SaleURL !== null) {
                for (x in p.SaleURL) {
                    var url = p.SaleURL[x];
                    if (url.search("taobao")>0 || url.search("tmall")>0)
                        $scope.data.taobaourl = url;
                    if (url.search("ebay")>0 )
                        $scope.data.ebayurl = url;
                }
            }
            
            var htmlintro = marked(p.Introduction);
            $('#introduction').html(htmlintro);
            var htmlspec = marked(p.Spec);
            $('#specs').html(htmlspec);
        });
    };
    
   
    $scope.reviewlater = function() {
        Enquires.reviewLater({'Id': $scope.deal.Id, 'NavName': $scope.deal.NavName, 'Name': $scope.deal.EnName,'CoverPhoto': $scope.deal.CoverPhoto, 'Price':$scope.deal.Price}, function(){
            $location.path('/myreviewboard');
        });
    };

}]);
