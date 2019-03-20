#include <ButtonDebounce.h>

const int LED_MODE_OFF = 0;
const int LED_MODE_ON = 1;
const int LED_MODE_GLOW = 2;
const int LED_MODE_BLINK = 3;

const int LED_MODE_BLINK_INTERVAL = 1000;
const int LED_MODE_GLOW_INTERVAL = 25;


const int PIN_ROTARY_A = 3; // Connected to CLK
const int PIN_ROTARY_B = 4; // Connected to DT
const int PIN_ROTARY_BTN = 5; // Connected to BTN
const int PIN_PUSHBUTTON = A1;
const int PIN_PUSHBUTTON_LED = 9;

int rotaryPinALast;
int rotaryBtnLast;
int rotaryPinAValue;
int rotaryButtonValue;

ButtonDebounce rotaryButton(PIN_ROTARY_BTN, 80);
ButtonDebounce pushButton(PIN_PUSHBUTTON, 80);


int rotaryRotation = 0;

int pushButtonValue;
int pushButtonLast;
int pushButtonLedMode = LED_MODE_BLINK;
int pushButtonLedValue = 0;
int pushButtonLedValueDirection = 1;

unsigned long pushButtonLedNextActionAt = 0;
unsigned long timeMillis;

void setup() {
  
 pinMode(PIN_ROTARY_A,INPUT);
 pinMode(PIN_ROTARY_B,INPUT);

 digitalWrite(PIN_PUSHBUTTON_LED, HIGH);

 rotaryPinALast = digitalRead(PIN_ROTARY_A);
 rotaryBtnLast = digitalRead(PIN_ROTARY_BTN);
 pushButtonLast = digitalRead(PIN_PUSHBUTTON);
 
 rotaryButton.setCallback(rotaryButtonChanged);
 pushButton.setCallback(pushButtonChanged);

 Serial.begin(9600);
}

void rotaryButtonChanged(const int state){
  if (state == 0) {
    Serial.print("D");
  }
}

void pushButtonChanged(const int state){
  if (state == 0) {
    Serial.print("P");
  }
}

void loop() {
  rotaryButton.update();
  pushButton.update();
  
  timeMillis = millis();

  /*
   * Read commands from serial
   */

   while(Serial.available()) {
    int cmd = Serial.read();
    Serial.println(cmd);
    switch (cmd) {
      case 66: // "B"
        pushButtonLedMode = LED_MODE_BLINK;
        break;
      case 71: // "G"
        pushButtonLedMode = LED_MODE_GLOW;
        break;
      case 79: // "O"
        pushButtonLedMode = LED_MODE_OFF;
        break;
    }
   }

  /*
   * Read rotary encoder
   */
  rotaryPinAValue = digitalRead(PIN_ROTARY_A);
  if (rotaryPinAValue != rotaryPinALast){ 
    // Means the knob is rotating
    // if the knob is rotating, we need to determine direction
    // We do that by reading pin B.
    if (digitalRead(PIN_ROTARY_B) != rotaryPinAValue) { 
      // Means pin A Changed first - We're Rotating Clockwise
      rotaryRotation++;
    } else {
      // Otherwise B changed first and we're moving CCW
      rotaryRotation--;
    }
    if (rotaryRotation == 2){
      Serial.print ("C");
      rotaryRotation = 0;
    } else if (rotaryRotation == -2) {
      Serial.print ("W");
      rotaryRotation = 0;
    }
  }
  rotaryPinALast = rotaryPinAValue;

  /*
   * Control LED
   */
  
  if (pushButtonLedMode == LED_MODE_OFF) {
    pushButtonLedValue = 0;
  } else if  (pushButtonLedMode == LED_MODE_ON) {
    pushButtonLedValue = 255;
  } else if (pushButtonLedMode == LED_MODE_BLINK) {
    if (timeMillis > pushButtonLedNextActionAt) {
      if (pushButtonLedValue == 0) {
        pushButtonLedValue = 1;
      } else {
        pushButtonLedValue = 0;
      }
      pushButtonLedNextActionAt = pushButtonLedNextActionAt + LED_MODE_BLINK_INTERVAL;
    }
  } else if (pushButtonLedMode == LED_MODE_GLOW) {
    if (timeMillis > pushButtonLedNextActionAt) {
      if (pushButtonLedValue < 32) {
        pushButtonLedValueDirection = 1;
      } else if (pushButtonLedValue > (255-32)) {
        pushButtonLedValueDirection = -1;
      }
      pushButtonLedValue += pushButtonLedValueDirection;
      pushButtonLedNextActionAt = pushButtonLedNextActionAt + LED_MODE_GLOW_INTERVAL;
    }
    
  }

  analogWrite(PIN_PUSHBUTTON_LED, pushButtonLedValue);

}
