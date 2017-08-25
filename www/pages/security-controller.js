/**
 * @class SecurityController
 * @classdesc
 * @ngInject
 */
function SecurityController($scope, SecurityService, ConfigLoader) {

  var DATE_FABRIC_FORMAT = 'yyyy-mm-dd'; // ISO

  var ctrl = this;

  ctrl.list = [];

  ctrl.accounts = ConfigLoader.getAllAccounts();

  /**
   *
   */
  ctrl.init = function(){
      // $scope.$on('chainblock', ctrl.reload);
      $scope.$on('chainblock-ch-'+ SecurityService.getChannelID(), ctrl.reload);
      ctrl.reload();
  }

  /**
   *
   */
  ctrl.reload = function(){
    ctrl.invokeInProgress = true;
    return SecurityService.list()
      .then(function(list){
        ctrl.list = list;
      })
      .finally(function(){
        ctrl.invokeInProgress = false;
      });
  }


  ctrl.newCalendarEntry = function(security){
    $scope.centrySecurity = security;
    $scope.centry = $scope.centry || {
      security: security.security,
      date: new Date()
    };
  }

  ctrl.sendCEntry = function(centry){
    ctrl.invokeInProgress = true;

    centry.date = formatDate(centry.date);
    return SecurityService.addCalendarEntry(centry)
      .then(function(){
        $scope.centry = null;
      })
      .finally(function(){
        ctrl.invokeInProgress = false;
      });
  }

  ctrl.sendSecurity = function(security){
    ctrl.invokeInProgress = true;
    return SecurityService.sendSecurity(security)
      .then(function(){
        $scope.security = null;
      })
      .finally(function(){
        ctrl.invokeInProgress = false;
      });
  }


  /**
   * Parse date in format dd/mm/yyyy
   * @param {string|Date} dateStr
   * @return {Date}
   */
  function formatDate(date){
    if(!date) return null;

    if(!(date instanceof Date)){
      // assumind date is a string: '1 August, 2017'
      // TODO: we shouldn't rely on this
      date = new Date(date);
    }
    return date.format(DATE_FABRIC_FORMAT);
  }



  ctrl.init();
}

angular.module('nsd.controller.security', ['nsd.service.security'])
.controller('SecurityController', SecurityController);