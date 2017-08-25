/**
 * @class BookController
 * @classdesc
 * @ngInject
 */
function BookController($scope, $q, BookService, ConfigLoader, DialogService, SecurityService) {

  var ctrl = this;

  ctrl.books = [];
  ctrl.securities = [];
  ctrl.accounts = ConfigLoader.getAllAccounts();

  /**
   *
   */
  ctrl.init = function(){
      $scope.$on('chainblock-ch-'+ BookService.getChannelID(), ctrl.reload);
      ctrl.reload();
  }

  /**
   *
   */
  ctrl.reload = function(){
    ctrl.invokeInProgress = true;

    return $q.all([

        SecurityService.list(SecurityService.STATUS_ACTIVE)
          .then(function(list){
            ctrl.securities = list;
          }),

        BookService.list()
          .then(function(list){
            // add 'org' and 'deponent' to the result, based on account+division
            list.forEach(function(item){
              item.org = ConfigLoader.getOrgByAccountDivision(item.balance.account, item.balance.division);
              item.deponent = (ConfigLoader.getAccount(item.org) || {}).dep;
            })
            ctrl.books = list;
          })

      ])
      .finally(function(){
        ctrl.invokeInProgress = false;
      });

  }



  /**
   * @param {Instruction} instruction
   */
  ctrl.showHistory = function(book){
    return BookService.history(book)
      .then(function(result){
        var scope = {history: result};
        return DialogService.dialog('book-history.html', scope);
      });
  }

  ctrl.sendBook = function(book){
    ctrl.invokeInProgress = true;
    return BookService.put(book)
      .finally(function(){
        ctrl.invokeInProgress = false;
      });
  }



  ctrl.init();
}

angular.module('nsd.controller.book', ['nsd.service.book'])
.controller('BookController', BookController);