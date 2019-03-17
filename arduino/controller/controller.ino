const int LED_MODE_OFF = 0;
const int LED_MODE_ON = 1;
const int LED_MODE_GLOW = 2;
const int LED_MODE_BLINK = 3;

const int PIN_ROTARY_A = 3; // Connected to CLK
const int PIN_ROTARY_B = 4; // Connected to DT
const int PIN_ROTARY_BTN = 5; // Connected to BTN
const int PIN_PUSHBUTTON = A1;
const int PIN_PUSHBUTTON_LED = 9;

int rotaryPinALast;
int rotaryBtnLast;
int rotaryPinAValue;
int rotaryButtonValue;

//boolean rotaryRotationIsClockwise;
int rotaryRotation = 0;

int pushButtonValue;
int pushButtonLast;
int btnLargeLedMode = LED_MODE_OFF;


void setup() {
  
 pinMode(PIN_ROTARY_A,INPUT);
 pinMode(PIN_ROTARY_B,INPUT);
 pinMode(PIN_ROTARY_BTN,INPUT_PULLUP);
 pinMode(PIN_PUSHBUTTON,INPUT_PULLUP);
 pinMode(PIN_PUSHBUTTON_LED, OUTPUT);

 digitalWrite(PIN_PUSHBUTTON_LED, HIGH);

 rotaryPinALast = digitalRead(PIN_ROTARY_A);
 rotaryBtnLast = digitalRead(PIN_ROTARY_BTN);
 pushButtonLast = digitalRead(PIN_PUSHBUTTON);
 
 Serial.begin(9600);
}

void loop() {

  /*
   * Read rotary push button 
   */
  rotaryButtonValue = digitalRead(PIN_ROTARY_BTN);
  if (rotaryButtonValue != rotaryBtnLast) {
    if (rotaryButtonValue == 0) {
      Serial.print ("Btn press\n");
    } else {
      Serial.print ("Btn release\n");
    }
    rotaryBtnLast = rotaryButtonValue;
  }

  /*
   * Read large push button
   */
  pushButtonValue = digitalRead(PIN_PUSHBUTTON);
  if (pushButtonValue != pushButtonLast) {
    if (pushButtonValue == 0) {
      Serial.print ("Btn large press\n");
    } else {
      Serial.print ("Btn large release\n");
    }
    pushButtonLast = pushButtonValue;
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
      //rotaryRotationIsClockwise = true;
      rotaryRotation++;
    } else {
      // Otherwise B changed first and we're moving CCW
      //rotaryRotationIsClockwise = false;
      rotaryRotation--;
    }
    if (rotaryRotation == 2){
      Serial.print ("Rotated: ");
      Serial.println ("clockwise");
      rotaryRotation = 0;
    } else if (rotaryRotation == -2) {
      Serial.print ("Rotated: ");
      Serial.println("counterclockwise");
      rotaryRotation = 0;
    }
  }
  rotaryPinALast = rotaryPinAValue;

  /*
   * Control LED
   */
  
  

}
