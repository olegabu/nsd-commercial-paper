(function(){

  /**
   * Here located nsd-specific directives, which are too small to place in a separate module
   */
  angular.module('nsd.directive.nsd', ['nsd.service.dialog'])
    // /**
    //  * @ngDirective role-show

    //  * @example <button confirm />
    //  * @example <button confirm="Are you sure?"/>
    //  */
    // .directive('confirm', function(DialogService, $timeout) {
    //   return {
    //     restrict:'A',
    //     priority: 50, // highter priority. default is 100 ?
    //     link: function(scope, elm, attrs, ctrl) {

    //       var confirmtext = attrs['confirm'];
    //       elm.on('click', function(e, data){
    //         e.stopImmediatePropagation();

    //         DialogService.confirm(confirmtext).then(function(){
    //           // elm.trigger(e);
    //           $timeout(function(){
    //             $(elm).trigger('click', data);
    //           });

    //         })
    //       })

    //     }
    //   };
    // })




})();
