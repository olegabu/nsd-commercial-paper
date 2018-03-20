/* globals angular */
/**
 * @class PositionsController
 * @classdesc
 * @ngInject
 */
function PositionsController($scope, $q, PositionsService, DialogService) {
  "use strict";

  var ctrl = this;

  ctrl.books = [];
  ctrl.moneys = [];

  /**
   *
   */
  ctrl.init = function(){
      $scope.$on('chainblock-ch-'+ PositionsService.getChannelID(), ctrl.reload);
      ctrl.reload();
  };

  /**
   *
   */
  ctrl.reload = function(){
    ctrl.invokeInProgress = true;

    return $q.all([


      PositionsService.list()
        .then(function(list){
          ctrl.books = list;
        })

    ])
      .finally(function(){
        ctrl.invokeInProgress = false;
      });

  };



  /**
   * @param {Book} book
   * @param {'paper'|'money'} type
   */
  ctrl.showHistory = function(book, type){
    return PositionsService.history(book)
      .then(function(result){
        var scope = {history: result, type: type};
        return DialogService.dialog('book-history.html', scope);
      });
  };


  ctrl.init();
}

angular.module('nsd.controller.positions', ['nsd.service.positions'])
.controller('PositionsController', PositionsController);