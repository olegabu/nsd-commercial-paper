/* globals angular */
/**
 *
 */
angular.module('nsd.controller', [
  'nsd.controller.main',
  'nsd.controller.login',
  'nsd.controller.book',
  'nsd.controller.positions',
  'nsd.controller.instructions',
  'nsd.controller.login',
  'nsd.controller.security'
]);

angular.module('nsd.service', [
  'nsd.service.api',
  'nsd.service.dialog',
  'nsd.service.channel',
  'nsd.service.socket',
  'nsd.service.user',
  'nsd.service.instructions',
  'nsd.service.book',
  'nsd.service.positions',
  'nsd.service.security'
]);

angular.module('nsd.app',[
   'ui.router',
   'ui.bootstrap',
   'ui.materialize',
   // 'ui.router.title',

   'LocalStorageModule',
   'jsonFormatter',
   'ngFileUpload',

   'nsd.config.env',
   'nsd.controller',
   'nsd.service',

   'nsd.directive.form',
   'nsd.directive.certificate',
   'nsd.directive.requsite',
   'nsd.directive.payment',
   'nsd.directive.blockchain',
   'nsd.directive.role',
   'nsd.directive.nsd',
   'pascalprecht.translate'
])
.config( function( $compileProvider ) {
  // https://stackoverflow.com/a/15769779/1269984

  $compileProvider.aHrefSanitizationWhitelist(/^\s*(https?|data):/);
  // Angular before v1.2 uses $compileProvider.urlSanitizationWhitelist(...)
})
.config(function($stateProvider) {
  // jshint: -W033

  /*
   Custom state options are:
      name:string      [<stateId>] - menu name
      title:string     [<stateId>] - page title name
      visible:boolean  [true]      - 'false' is hide element from menu. setting abstract:true also hide it
      roles:string                 - array of of nsd|issuer|investor
      (not implemented) default:boolean  [false]     - navigate to this item by default

  */
  $stateProvider
    .state('app', {
      url: '/',
      abstract:true,
      templateUrl: 'app.html',
      resolve: {
        // $title: function() { return 'Home'; }
        _config: function(ConfigLoader){ return ConfigLoader.load(); }
      }
    })
    .state('app.login', {
      url: 'login',
      templateUrl: 'pages/login.html',
      controller: 'LoginController',
      controllerAs: 'ctl',
      data:{
        name: false,
        visible: false,
        roles:'*'
      }
    })

    .state('app.book', {
      url: 'book',
      templateUrl  : 'pages/book.html',
      controller   : 'BookController',
      controllerAs : 'ctl',
      data:{
        name: 'BOOK_MENU',
        title: false,
        roles:'nsd'
      }
    })
    .state('app.positions', {
      url: 'positions',
      templateUrl  : 'pages/positions.html',
      controller   : 'PositionsController',
      controllerAs : 'ctl',
      data:{
        name: 'POSITIONS_MENU',
        title: false,
        roles: ['issuer', 'investor']
      }
    })
    .state('app.security', {
      url: 'security',
      templateUrl  : 'pages/security.html',
      controller   : 'SecurityController',
      controllerAs : 'ctl',
      data:{
        name: 'SECURITY_MASTER_MENU',
        roles:'*'
      }
    })
    .state('app.instructions', {
      url: 'instructions',
      templateUrl  : 'pages/instructions.html',
      controller   : 'InstructionsController',
      controllerAs : 'ctl',
      data:{
        name: 'INSTRUCTION_MENU',
        roles:'*',
        default: true
      }
    })
    // .state('app.explorer', {
    //   url: '/admin',
    //   templateUrl  : 'pages/explorer.html',
    //   controller   : 'ExplorerController',
    //   controllerAs : 'ctl',
    //   data:{
    //     absolute: true,
    //     name: 'Explorer',
    //     roles:'*'
    //   }
    // })

})

.run(function(env, $log){
  if (!env.role) {
    $log.warn('Client role not set');
    env.role = '*';
  }
})

// THIS method should be called BEFORE navigateDefault()
.run(function(UserService, $rootScope, $state, $log, $window){

  //
  var loginState = 'app.login';

  // https://github.com/angular-ui/ui-router/wiki#state-change-events
  $rootScope.$on('$stateChangeStart',  function(event, toState, toParams, fromState, fromParams, options){
    // console.log('$stateChangeStart', event, toState, toParams);
    toState.data = toState.data || {};

    // check access
    var isAllowed = UserService.canAccess(toState);
    var isLoginState = toState.name == loginState;

    if ( isLoginState && !isAllowed){
      $log.warn('login state cannot be forbidden');
      isAllowed = true;
    }

    $log.debug('$stateChangeStart access: state - %s, allowed - %s', toState.name, isAllowed);
    // prevent navigation to forbidden pages
    if ( !UserService.isAuthorized() && !isAllowed && !isLoginState){
      event.preventDefault(); // transitionTo() promise will be rejected with a 'transition prevented' error
      if(fromState.name == ""){
        // just enter the page - redirect to login page
        $log.debug('Redirect to login page');
        goLogin();
        return
      }else{
        // we are at some page and try to go to forbidden one.
        // just ignore this attempt
      }
    }

    // if(toState.data.absolute){
    //   event.preventDefault(); // transitionTo() promise will be rejected with a 'transition prevented' error
    //   $window.location = toState.url;
    //   return;
    // }

  });

  // set state data to root scope
  $rootScope.$on('$stateChangeSuccess',  function(event, toState, toParams, fromState, fromParams, options){
    $rootScope.$state = toState;
    $rootScope.$stateParams = toParams;
  });

  /**
   *
   */
  function goLogin(){
    $state.go(loginState);
  }
})

// instead of: $urlRouterProvider.otherwise('/default');
.run(function navigateDefault($state, $log, $rootScope){

  var defaultState = getDefaultState();
  if(!defaultState){
    $log.warn('No default state set. Please, mark any state as default by setting "data:{ default:true }"');
  }
  $rootScope.stateDefault = defaultState; // TODO: remove?

  // instead of: $urlRouterProvider.otherwise('/default');
  if($state.current.name == "" && defaultState){
    $state.go(defaultState.name);
  }

  /**
   * @return {State}
   */
  function getDefaultState(){
    var states = $state.get()||[];
    for (var i = states.length - 1; i >= 0; i--) {
      if( states[i].data && states[i].data.default === true){
        return states[i];
      }
    }
    return null;
  }

})


/**
 *
 */
.config(function($httpProvider) {
  $httpProvider.interceptors.push('bearerAuthIntercepter');
})

/**
 *
 */
.config(['$translateProvider', function ($translateProvider) {
    $translateProvider.useStaticFilesLoader({
        prefix: 'i18n/locale-',
        suffix: '.json'
    });
    $translateProvider.preferredLanguage('ru');
    // $translateProvider.preferredLanguage('en');
    $translateProvider.useSanitizeValueStrategy(null); // escape|sanitize
    // $translateProvider.usePostCompiling(true);
}])

/**
 * inject 'X-Requested-With' header
 * inject 'Authorization: Bearer' token
 */
.factory('bearerAuthIntercepter', function($rootScope){
    return {
        request: function(config) {
            config.headers['X-Requested-With'] = 'XMLHttpRequest'; // make ajax request visible among the others
            config.withCredentials = true;

            // $rootScope._tokenInfo is set in UserService
            if($rootScope._tokenInfo){
              config.headers['Authorization'] = 'Bearer '+$rootScope._tokenInfo.token;
            }
            return config;
        },

        // throws error, so '$exceptionHandler' service will caught it
        requestError:function(rejection){
          throw rejection;
        },
        responseError:function(rejection){
          throw rejection;
        }
    };

})





/**
 * load config from remote endpoint
 * @deprecated: use environment service
 * @ngInject
 */
.service('ConfigLoader', function(ApiService, $rootScope) {
    "use strict";

  /**
   * @typedef {object} FabricConfig
   * @property {string} org - your org ID
   * @property {string[]} endorsers - org ID of root endorsers
   *
   * @property {object} network-config
   * @property {string} network-config.role
   * @property {string} network-config.dep
   * @property {object} network-config.acc
   *
   * @property {object} account-config
   * @property {object} account-config.orderer
   * @property {object} account-config.<org>.acc
   * @property {object} account-config.<org>.bic
   * @property {string} account-config.<org>.dep
   * @property {string} account-config.<org>.role
   */


  /** @type {Promise<FabricConfig>} */
    var configPromise;
    var _config = null;

    function _resolveConfig(){
      if( !configPromise ){
        configPromise = ApiService.getConfig()
          .then(function(config){
            $rootScope._config = config;
            _config = config;
            _extendConfig();
            _extendAccountConfig();
            console.log('ConfigLoader - got config:', config); // jshint ignore:line
            return config;
          });
      }
      return configPromise;
    }


    window.dev = function(isDev) {
      _config.dev = (typeof isDev == 'undefined' ? true : isDev);
      return _config.dev;
    };


    /**
     * add getOrgs() to netConfig
     * add getPeers(orgId:string) to netConfig

     * add id to org info

     * add id to peer info
     * add org to peer info
     * add host to peer info
     */
    function _extendConfig(){
      var netConfig = _config['network-config'];

      Object.keys(netConfig)
        .filter(function(key){ return key !== 'orderer'; })
        .forEach(function(orgId){

          // add org.id
          netConfig[orgId].id = orgId;

          var orgConfig = netConfig[orgId] || {};

          // add peers stuff
          Object.keys(orgConfig)
            .filter(function(key){ return key.startsWith('peer') })
            .forEach(function(peerId){
              orgConfig[peerId].id   = peerId;
              orgConfig[peerId].host = getHost(orgConfig[peerId].requests);
              orgConfig[peerId].org  = orgId;
            });

        });
    }

  function _extendAccountConfig(){
    var accConfig = _config['account-config'];

    Object.keys(accConfig)
      .forEach(function(orgId){

        accConfig[orgId].bic = {};
        Object.keys(accConfig[orgId].acc).forEach(function(account){
          if (account.length > 12) {
            // assume it's bic
            accConfig[orgId].bic[account] = accConfig[orgId].acc[account];
            delete accConfig[orgId].acc[account];
          }
        });

      });
  }

    function getPeers(orgId){
        var netConfig = _config['network-config'];
        var orgConfig = netConfig[orgId]||{};

        return Object.keys(orgConfig)
          .filter(function(key){ return key.startsWith('peer')})
          .map(function(key){ return orgConfig[key]; });
    }

    function getOrgs(){
      var netConfig = _config['network-config'];

      return Object.keys(netConfig)
        .filter(function(key){ return key != 'orderer'})
        .map(function(key){ return netConfig[key]; });
    }

    /**
     *
     */
    function getHost(address){
      //                             1111       222222
      var m = (address||"").match(/^(\w+:)?\/\/([^\/]+)/) || [];
      return m[2];
    }

    /**
     * @param {string} orgID
     */
    function getAccount(orgID){
      var accountConfig = _config['account-config'] || {};
      return accountConfig[orgID];
    }

    /**
     *
     */
    function getAccounts() {
      return _config['account-config'] || {};
    }

    /**
     * @return {Object<Array<string>>} key-value of account => [divisions]
     * @deprecated
     */
    function getAllAccounts(){
      var accountConfig = _config['account-config'] || {};
      return Object.keys(accountConfig).reduce(function(result, orgID){

        Object.keys( accountConfig[orgID].acc ).forEach(function(account){
          result[account] = angular.copy(accountConfig[orgID].acc[account]);
          result[account].org = orgID;
        });
        return result;
      }, {});
    }

    /**
     * @return {Object<Array<string>>} key-value of account => [divisions]
     * @deprecated
     */
    function getAllBics(){
      var accountConfig = _config['account-config'] || {};
      return Object.keys(accountConfig).reduce(function(result, orgID){

        Object.keys( accountConfig[orgID].bic ).forEach(function(account){
          result[account] = angular.copy(accountConfig[orgID].bic[account]);
          result[account].org = orgID;
        });
        return result;
      }, {});
    }

    function getOrg(){
      return _config.org;
    }


    /**
     * get organosation ID by deponent code (1 to 1 matching)
     * @param  {string} depCode
     * @return {string} orgID
     */
    function getOrgByDepcode(depCode){
      var accountConfig = _config['account-config'];
      // looking for second participant
      for(var org in accountConfig){
        if(accountConfig.hasOwnProperty(org)){
          if(accountConfig[org].dep == depCode){
            return org;
            // break;
          }
        }
      }
      return null;
    }

    /**
     *
     */
    function getOrgPeerIds(org){
      var netConfig = _config['network-config'];
      if (!netConfig[org]) {
        throw new Error('Organisation not found: ' + org);
      }
      return Object.keys(netConfig[org])
        .filter(function(key){ return key.startsWith('peer'); });
    }

    /**
     *
     */
    function getOrgByAccountDivision(account, division){
      var accountConfig = _config['account-config'] || {};
      var orgArr = Object.keys(accountConfig);
      for (var i = orgArr.length - 1; i >= 0; i--) {
        var orgID = orgArr[i];
        if( accountConfig[orgID].acc[account] && accountConfig[orgID].acc[account].indexOf(division)>=0 ) {
          return orgID;
        }
        // ignore division for bic
        if( accountConfig[orgID].bic[account] ) {
          return orgID;
        }
      }
      return null;
    }


    /////////
    return {
      load  :   _resolveConfig,
      ready :   _resolveConfig,
      getOrg     : getOrg,
      getAccount : getAccount,
      getAccounts: getAccounts,
      getPeers   : getPeers,
      getOrgs    : getOrgs,
      getAllAccounts : getAllAccounts,
      getAllBics : getAllBics,

      getOrgByAccountDivision : getOrgByAccountDivision,
      getOrgByDepcode : getOrgByDepcode,
      getOrgPeerIds   : getOrgPeerIds,
      get:function(){ return _config; }
    };
})

/**
 * @ngInject
 */
.run(function(ConfigLoader){
  ConfigLoader.load();
})


.factory('$exceptionHandler', function ($window) {
    $window.onunhandledrejection = function(e) {
        // console.warn('onunhandledrejection', e);
        e = e || {};
        onError(e);
    };

    function onError(exception){
        // filter network 403 errors
        if (exception.status !== 403 ){
            globalErrorHandler(exception);
        }
    }

    return onError;
})



/**
 *
 */
function globalErrorHandler(e){
  console.warn('globalErrorHandler', e);
  e = e || {};
  if(typeof e == "string"){
    e = {message:e};
  }
  e.data = e.data || {};

  var statusMsg = e.status ? 'Error' + (e.status != -1?' '+e.status:'') + ': ' + (e.statusText||(e.status==-1?"Connection refused":null)||"Unknown") : null;
  var reason = (e.data ? e.data.message : null) || e.reason || e.message || statusMsg || e || 'Unknown error';
  Materialize.toast(reason, 4000, 'mytoast red') // 4000 is the duration of the toast
}