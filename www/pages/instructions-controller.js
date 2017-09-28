/**
 * @class InstructionsController
 * @classdesc
 * @ngInject
 */
function InstructionsController($scope, $q, $filter, InstructionService, BookService, UserService, DialogService, ConfigLoader /*, SocketService*/) {

  var ctrl = this;
  ctrl.list = [];
  ctrl.redeemList = [];

  // var DATE_INPUT_FORMAT = 'dd/mm/yyyy';
  var DATE_FABRIC_FORMAT = 'yyyy-mm-dd'; // ISO
  var TRANSFER_SIDE_TRANSFERER = 'transferer';
  var TRANSFER_SIDE_RECEIVER = 'receiver';
  var NSD_ROLE = 'nsd';


  ctrl.org = ConfigLoader.get().org;
  ctrl.account = ConfigLoader.getAccount(ctrl.org);

  // ConfigLoader.getAccount(orgID)
  ctrl.accountFrom = null;
  ctrl.accountTo   = null;


  /**
   *
   */
  ctrl.init = function(){
      $scope.$on('chainblock', function(e, block){
        if( InstructionService.isBilateralChannel(block.getChannel()) || block.getChannel() == BookService.getChannelID()){
          ctrl.reload();
        }
      });
      ctrl.reload();
  }

  /**
   *
   */
  ctrl.reload = function(){
    ctrl.invokeInProgress = true;
    return $q.all([
      InstructionService.listAll()
        .then(function(list){
          ctrl.list = list;
        }),

      UserService.getOrgRole() !== NSD_ROLE
      ? $q.resolve()
      : BookService.redeemHistory()
        .then(function(redeemList){
          ctrl.redeemList = redeemList;
        })
    ])
    .finally(function(){
      ctrl.invokeInProgress = false;
    });
  }


  ctrl.isTransferer = function(instruction){
    var acc = Object.keys(ctrl.account.acc);
    return acc.indexOf(instruction.transferer.account) > -1;
  }

  ctrl.isReceiver = function(instruction){
    var acc = Object.keys(ctrl.account.acc);
    return acc.indexOf(instruction.receiver.account) > -1;
  }


  ctrl.isInitiator = function(instruction){
    return instruction.initiator=='transferer' ? ctrl.isTransferer(instruction) : ctrl.isReceiver(instruction);
  }

  ctrl.isAdmin = function(){
    return ctrl.org === NSD_ROLE;
  }


  /**
   *
   * @param inst instruction
   * @param type instruction type
   * @returns {boolean} true if instruction should be displayed
   */
  ctrl.showInstruction = function(inst, type) {
    var acc = Object.keys(ctrl.account.acc);
    return ctrl.org === NSD_ROLE ||
        (type === TRANSFER_SIDE_TRANSFERER && acc.indexOf(inst.transferer.account) > -1) ||
        (type === TRANSFER_SIDE_RECEIVER && acc.indexOf(inst.receiver.account) > -1);
        // (acc.indexOf(inst.transferer.account) > -1) || (acc.indexOf(inst.receiver.account) > -1);
  }

  /**
   * Displays reason based on role
   *
   * @param inst instruction
   * @param key object key that should be compared
   */
  ctrl.showReason = function(inst, key) {
    var curDep = inst[key];
    return ctrl.org === NSD_ROLE || curDep === ctrl.account.dep;
  }

  /**
   * @return {Instruction}
   */
  ctrl._getDefaultInstruction = function(transferSide, opponentID){
    var orgID = ctrl.org;
    return {
      deponentFrom: ctrl._getDeponentCode(transferSide == TRANSFER_SIDE_TRANSFERER ? orgID : opponentID),
      deponentTo:   ctrl._getDeponentCode(transferSide == TRANSFER_SIDE_RECEIVER ? orgID : opponentID),

      initiator: transferSide,
      // quantity: 0, // TODO: cause ui bug with overlapping label and input field with value
      tradeDate    : new Date(),
      instructionDate : new Date()
    };
  }

  ctrl._getDeponentCode = function(orgID){
    if(orgID === ctrl.org) {
      return ctrl.account.dep;
    }
    var account = ConfigLoader.getAccount(orgID) || {};
    return account.dep;
  }

  /**
   *
   */
  ctrl.getStatusClass = function(status){
    switch(status){
      case 'matched' : return 'deep-purple-text';
      case 'declined': return 'red-text darken-4';
      case 'executed': return 'green-text darken-4';
      case 'canceled': return 'grey-text';
      default: return '';
    }
  }



  ctrl.cancelInstruction = function(instruction){

    return DialogService.confirm( $filter('translate')('CANCEL_INSTRUCTION_PROMPT').replace('%s', instruction.deponentFrom).replace('%s', instruction.deponentTo), {yesKlass:'red white-text'})
      .then(function(isConfirmed){
        if(isConfirmed){
          ctrl.invokeInProgress = true;
          return InstructionService.cancelInstruction(instruction)
            .finally(function(){
              ctrl.invokeInProgress = false;
            });
        }
      })

  }


  /**
   *
   */
  ctrl.newInstructionTransfer = function(transferSide, _channel){
    if(!$scope.inst || $scope.inst.initiator != transferSide){
        // preset values

        var opponentOrgID = ctrl._getOrgIDByChannel(_channel);
        $scope.inst = ctrl._getDefaultInstruction(transferSide, opponentOrgID);
        $scope.formInstruction.$setPristine();

        // preset
        ctrl._fillAccount(transferSide, opponentOrgID);
    }
  };

  ctrl._fillAccount = function(transferSide, opponentID){
    if(transferSide == TRANSFER_SIDE_TRANSFERER){
      ctrl.accountFrom = ConfigLoader.getAccount(ctrl.org);
      ctrl.accountTo = opponentID ? ConfigLoader.getAccount(opponentID) : null;
    } else {
      ctrl.accountFrom = opponentID ? ConfigLoader.getAccount(opponentID) : null;
      ctrl.accountTo = ConfigLoader.getAccount(ctrl.org);
    }
  };



  /**
   *
   */
  ctrl._getOrgIDByChannel = function(channelID){
    if(!channelID) return null;
    return channelID.split('-').filter(function(org){ return org != ctrl.org; })[0];
  }

  /**
   *
   */
  ctrl.sendInstruction = function(instruction){
    $scope.inst = null;

    // FIXME here date can come in two different formats:
    //  Date object when we change form value
    //  String (like '1 August, 2017') when we not change form value
    // Now we use formatDate() to transform both of it into ISO
    instruction.tradeDate        = formatDate(instruction.tradeDate);
    instruction.instructionDate  = formatDate(instruction.instructionDate);
    if(instruction.reason && instruction.reason.created){
      instruction.reason.created   = formatDate(instruction.reason.created);
    }

    var p;
    switch(instruction.initiator){
      case TRANSFER_SIDE_TRANSFERER:
        p = InstructionService.transfer(instruction);
        break;
      case TRANSFER_SIDE_RECEIVER:
        p = InstructionService.receive(instruction);
        break;
      default:
        throw new Error('Unknpown transfer side: ' + instruction.initiator);
    }


    ctrl.invokeInProgress = true;
    return p.finally(function(){
      ctrl.invokeInProgress = false;
    });
  };

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


  /**
   *
   */
  ctrl.newRedemption = function(){
    $scope.redemption = $scope.redemption || ctrl._getDefaultRedemption();
  }
  /**
   * @return {Redemption}
   */
  ctrl._getDefaultRedemption = function(){
    return {
      reason:{
        created   : new Date()//.format(DATE_INPUT_FORMAT)
      }
    };
  }

  /**
   * @param {Redemption} redemption
   */
  ctrl.sendRedemption = function(redemption){
    return DialogService.confirm( $filter('translate')('REDEEM_INSTRUCTION_PROMPT').replace('%s', redemption.security), {yesKlass:'red white-text'})
      .then(function(isConfirmed){
        if(isConfirmed){

          ctrl.invokeInProgress = true;
          return BookService.redeem(redemption)
            .finally(function(){
              ctrl.invokeInProgress = false;
            });
        }
      })
      .then(function(){
        $scope.redemption = null;
      });
  }


  /**
   * @param {Instruction} instruction
   */
  ctrl.showHistory = function(instruction){
    return InstructionService.history(instruction)
      .then(function(result){
        var scope = {history: result, getStatusClass: ctrl.getStatusClass, showReason: ctrl.showReason};
        return DialogService.dialog('balance-history.html', scope);
      });
  }

  ctrl.showABPrefill = function(transferSide){
    var orglc= (''+UserService.getOrg()).toLowerCase();
      return ( orglc === 'megafon' && transferSide == 'transferer')
          || ( orglc === 'raiffeisen' && transferSide == 'receiver');
  }

  ctrl.getABStub = function(transferSide){
    var accountConfig = ConfigLoader.get()['account-config'];
    var orgFrom = 'megafon';
    var orgTo   = 'raiffeisen';
    return {
      deponentFrom: accountConfig[orgFrom].dep,
      deponentTo:   accountConfig[orgTo].dep,

      security:'RU000A0JWGG3',
      transferer:{
        account : Object.keys(accountConfig[orgFrom].acc)[0],
        division: accountConfig[orgFrom].acc[ Object.keys(accountConfig[orgFrom].acc)[0] ][0]
      },
      receiver:{
        account : Object.keys(accountConfig[orgTo].acc)[0],
        division: accountConfig[orgTo].acc[ Object.keys(accountConfig[orgTo].acc)[0] ][0]
      },
      initiator: transferSide,
      quantity: 1,
      reference: 'test',
      memberInstructionId:123,
      tradeDate    : new Date(),
      instructionDate : new Date()
    };
  }


  ctrl.getACStub = function(transferSide){
    return {
      deponentFrom: 'CA9861913023',
      deponentTo:   'NL0000729408',

      security:'RU0DLTMFONCB',
      transferer:{
        account: "AC0689654902",
        division: "87680000045800005",
      },
      receiver:{
        account: "YN0927654908",
        division: "37800007360900016",
      },
      initiator: transferSide,
      quantity: 1,
      reference: 'test',
      memberInstructionId:123,
      tradeDate    : new Date(),
      instructionDate : new Date()
    };
  }




  //////////////

  // INIT
  ctrl.init();

}

angular.module('nsd.controller.instructions', ['nsd.service.instructions'])
.controller('InstructionsController', InstructionsController);