/**
 * @class PositionsController
 * @classdesc
 * @ngInject
 */
function PositionsController($scope, PositionsService, ConfigLoader, DialogService) {

  var ctrl = this;

  ctrl.list = [];

  /**
   *
   */
  ctrl.init = function(){
      $scope.$on('chainblock-ch-'+ PositionsService.getChannelID(), ctrl.reload);
      ctrl.reload();
  }

  /**
   *
   */
  ctrl.reload = function(){
    ctrl.invokeInProgress = true;
    return PositionsService.list()
      .then(function(list){
        // add 'org' and 'deponent' to the result, based on account+division
        list.forEach(function(item){
          item.org = ConfigLoader.getOrgByAccountDivision(item.balance.account, item.balance.division);
          item.deponent = (ConfigLoader.getAccount(item.org) || {}).dep;
        })
        ctrl.list = list;
      })
      .finally(function(){
        ctrl.invokeInProgress = false;
      });
  }


  /**
   * @param {Instruction} instruction
   */
  ctrl.showHistory = function(book){
    return PositionsService.history(book)
      .then(function(result){
        var scope = {history: result};
        return DialogService.dialog('book-history.html', scope);
      });
  }


  ctrl.init();
}

angular.module('nsd.controller.positions', ['nsd.service.positions'])
.controller('PositionsController', PositionsController);