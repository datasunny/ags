angular.module('featen.report').controller('ReportsController', ['$scope', '$routeParams', '$location', 'Global', 'Reports', function($scope, $routeParams, $location, Global, Reports) {
        var chartdata;
        
        $scope.getNewCustomerChart1 = function(){
        	$scope.getNewCustomerChart(1);
        };
        
        
        $scope.getNewCustomerChart = function(timeframe) {
            var conds = [];
            conds.push("timeframe=" + timeframe);
            conds.push("type=NewCustomers");


            var cond = conds.join("&");

            Reports.getdata(cond, function(ps) {
                var v = [];
                var now = (new Date());
                now.setHours(0);
                now.setMinutes(0);
                now.setSeconds(0);
                now.setMilliseconds(0);

                var d = new Date(now.getTime());
                switch (ps.Timeframe) {
                    case "0": //7days
                        for (d.setDate(now.getDate() - 7); d <= now; d.setDate(d.getDate() + 1)) {
                            v.push([d.getTime(), 0]);
                        }
                        break;
                    case "1": //30days
                        for (d.setDate(now.getDate() - 30); d <= now; d.setDate(d.getDate() + 1)) {
                            v.push([d.getTime(), 0]);
                        }
                        break;
                    case "2": //180days
                        for (d.setDate(now.getDate() - 180); d <= now; d.setDate(d.getDate() + 1)) {
                            v.push([d.getTime(), 0]);
                        }
                        break;
                    case "3": //all
                        var dd = new Date(ps.Xvalues[0]);
                        dd.setHours(0);
                        
                        for (d=dd; d <= now; d.setDate(d.getDate() + 1)) {
                            v.push([d.getTime(), 0]);
                        }
                        break;
                }

                if (ps.Xvalues !== null) {
                    for (var i = 0; i < ps.Xvalues.length; i++) {
                        var dd = new Date(ps.Xvalues[i]);
                        dd.setHours(0);
                        var dv = dd.getTime();
                        for (var j = 0; j < v.length; j++) {
                            if (v[j][0] === dv) {
                                v[j][1] = ps.Yvalues[i];
                            }
                        }
                        //[(new Date(ps.Xvalues[i])).getTime(), ps.Yvalues[i]];
                    }
                }
                chartdata = [{
                        //*values: [[1136005200000, 12], [1138683600000, 3], [1141102800000, 5], [1143781200000, 18], [1151640000000, 100]],
                        values: v,
                        key: ps.Type,
                        color: "#ff7f0e"
                    }].map(function(series) {
                    series.values = series.values.map(function(d) {
                        return {x: d[0], y: d[1]};
                    });
                    return series;
                });
                genChart();
            });


        };




        function genChart() {
            nv.addGraph(function() {
                var width = 600;
                var height = 300;
                var chart = nv.models.lineChart()
                        .x(function(d, i) {
                            return i;
                        });

                chart.xAxis
                        .axisLabel('Date')
                        .tickFormat(function(d) {
                            var dx = chartdata[0].values[d] && chartdata[0].values[d].x || 0;
                            return dx ? d3.time.format('%x')(new Date(dx)) : '';
                        });

                chart.yAxis
                        .axisLabel('')
                        .tickFormat(d3.format('f'));
                resizeChart();

                d3.select('#chart1 svg')
                        .attr('perserveAspectRatio', 'xMinYMid')
                        .attr('width', width)
                        .attr('height', height)
                        .datum(chartdata)
                        .transition().duration(500)
                        .call(chart);

                nv.utils.windowResize(resizeChart);
                function resizeChart() {
                    var container = d3.select('#chart1');
                    var svg = container.select('svg');


                    // resize based on container's width
                    var aspect = 2;
                    var targetWidth = parseInt(container.style('width'));
                    svg.attr("width", targetWidth);
                    svg.attr("height", Math.round(targetWidth / aspect));

                }
                return chart;
            });
        }


    }]);
