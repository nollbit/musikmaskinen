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

int pushButtonValue;
int pushButtonLast;
int btnLargeLedMode = LED_MODE_OFF;


boolean bCW;
void setup() {
  
 pinMode(PIN_ROTARY_A,INPUT);
 pinMode(PIN_ROTARY_B,INPUT);
 pinMode(PIN_ROTARY_BTN,INPUT_PULLUP);
 pinMode(PIN_PUSHBUTTON,INPUT_PULLUP);
 pinMode(PIN_PUSHBUTTON_LED, OUTPUT);

 rotaryPinALast = digitalRead(PIN_ROTARY_A);
 rotaryBtnLast = digitalRead(PIN_ROTARY_BTN);
 pushButtonLast = digitalRead(PIN_PUSHBUTTON);
 
 Serial.begin(9600);
}

void loop() {

  rotaryButtonValue = digitalRead(PIN_ROTARY_BTN);
  if (rotaryButtonValue != rotaryBtnLast) {
    if (rotaryButtonValue == 0) {
      Serial.print ("Btn press\n");
    } else {
      Serial.print ("Btn release\n");
    }
    rotaryBtnLast = rotaryButtonValue;
  }

  pushButtonValue = digitalRead(PIN_PUSHBUTTON);
  if (pushButtonValue != pushButtonLast) {
    if (pushButtonValue == 0) {
      Serial.print ("Btn large press\n");
    } else {
      Serial.print ("Btn large release\n");
    }
    pushButtonLast = pushButtonValue;
  }

  rotaryPinAValue = digitalRead(PIN_ROTARY_A);
  if (rotaryPinAValue != rotaryPinALast){ 
    // Means the knob is rotating
    // if the knob is rotating, we need to determine direction
    // We do that by reading pin B.
    if (digitalRead(PIN_ROTARY_B) != rotaryPinAValue) { 
      // Means pin A Changed first - We're Rotating Clockwise
      bCW = true;
    } else {
      // Otherwise B changed first and we're moving CCW
      bCW = false;
    }
    Serial.print ("Rotated: ");
    if (bCW){
      Serial.println ("clockwise");
    }else{
      Serial.println("counterclockwise");
    }
  }
  rotaryPinALast = rotaryPinAValue;
}
