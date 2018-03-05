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
  ctrl.accounts = ConfigLoader.getAllAccounts();

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
            ctrl.securities = list;
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
   */
  ctrl.showHistory = function(book){
    return BookService.history(book)
      .then(function(result){
        var scope = {history: result};
        return DialogService.dialog('book-history.html', scope);
      });
  };

  /**
   * prepare book for create/update book
   * @param {Book} [book]
   */
  ctrl.newBook = function(book){

  };

  /**
   * prepare book for create/update book
   * @param {Book} [book]
   * @return {number}
   */
  ctrl.getBookBalance = function(book){
    if(book && book.balance) {
      for (var i = ctrl.books.length - 1; i >= 0; i--) {
        var b = ctrl.books[i];

        if (b.balance.account === book.balance.account && b.balance.division === book.balance.division) {
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
      })
      .finally(function(){
        ctrl.invokeInProgress = false;
      });
  };



  ctrl.init();
}

angular.module('nsd.controller.book', ['nsd.service.book'])
.controller('BookController', BookController);