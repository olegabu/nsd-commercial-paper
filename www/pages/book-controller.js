/* globals angular */
/**
 * @class BookController
 * @classdesc
 * @ngInject
 */
function BookController($scope, $q, BookService, ConfigLoader, DialogService, SecurityService) {
  "use strict";

  var ctrl = this;

  ctrl.books = [];
  ctrl.securities = [];
  ctrl.moneys = [];
  ctrl.accounts = ConfigLoader.getAllAccounts();
  ctrl.bics = ConfigLoader.getAllBics();

  /**
   *
   */
  ctrl.init = function(){
      $scope.$on('chainblock-ch-'+ BookService.getChannelID(), ctrl.reload);
      ctrl.reload();
  };

  /**
   *
   */
  ctrl.reload = function(){
    ctrl.invokeInProgress = true;

    return $q.all([

        SecurityService.list(SecurityService.STATUS_ACTIVE)
          .then(function(list){
            ctrl.securities = list.filter(function(security){
              return security.type === SecurityService.TYPE_PAPER;
            });
            ctrl.moneys = list.filter(function(security){
              return security.type === SecurityService.TYPE_MONEY;
            });
          }),

        BookService.list()
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
    return BookService.history(book)
      .then(function(result){
        var scope = {history: result, type: type};
        return DialogService.dialog('book-history.html', scope);
      });
  };

  /**
   * prepare book for create/update book
   */
  ctrl.newBook = function(type) {
    $scope.book = $scope.book || {};
    $scope.book.type = type;
    if ($scope.book.type === 'money') {
      $scope.book.balance = $scope.book.balance || {};
      $scope.book.balance.division = '';
      $scope.book.security = 'RUB';
    }
  };

  /**
   * prepare book for create/update book
   * @param {Book} [book]
   * @return {number}
   */
  ctrl.getBookBalance = function(book){
    if(book && book.balance) {
      for (var i = ctrl.books.length - 1; i >= 0; i--) {
        var b = /** @type {Book} */ ctrl.books[i];

        // jshint -W014
        if (b.balance.account === book.balance.account
          && b.balance.division === book.balance.division
          && b.security === book.security) {
          return b.quantity;
        }
      }
    }
    return 0;
  };

  ctrl.addBook = function(book){
    ctrl.invokeInProgress = true;
    return BookService.put(book)
      .then(function(){
        $scope.book = null;
        $scope.bookForm.$setPristine();
        $scope.moneyForm.$setPristine();
      })
      .finally(function(){
        ctrl.invokeInProgress = false;
      });
  };



  ctrl.init();
}

angular.module('nsd.controller.book', ['nsd.service.book'])
.controller('BookController', BookController);